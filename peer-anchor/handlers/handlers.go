package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/protobuf/types/known/timestamppb"

	grpcclient "peer-anchor/grpc-client"
	"peer-anchor/models"
	"peer-anchor/proto"
)

type SupplierDTO struct {
	SupplierId string  `json:"supplierId"`
	Name       string  `json:"name"`
	Amount     float64 `json:"amount"`
	Status     string  `json:"status"`
}

type Handler struct {
	db            *mongo.Database
	ordererClient *grpcclient.OrdererClient
	peerID        string
}

func generateEventID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// getPreviousBlockHash returns the hash of the previous block
func (h *Handler) getPreviousBlockHash() string {
	var lastBlock map[string]interface{}
	err := h.db.Collection("blocks").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"blockNumber": -1})).Decode(&lastBlock)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "genesis" // Genesis block has no previous hash
		}
		return "" // Return empty on error
	}

	if hash, ok := lastBlock["hash"].(string); ok {
		return hash
	}
	return ""
}

// getPreviousBlockNumber returns the number of the previous block
func (h *Handler) getPreviousBlockNumber() int64 {
	var lastBlock map[string]interface{}
	err := h.db.Collection("blocks").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"blockNumber": -1})).Decode(&lastBlock)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0 // No previous block
		}
		return 0 // Return 0 on error
	}

	// Handle different numeric types that MongoDB might return
	if blockNum, ok := lastBlock["blockNumber"].(int64); ok {
		return blockNum
	}
	if blockNum, ok := lastBlock["blockNumber"].(int32); ok {
		return int64(blockNum)
	}
	if blockNum, ok := lastBlock["blockNumber"].(int); ok {
		return int64(blockNum)
	}
	return 0
}

// calculateBlockHash calculates SHA256 hash of block data
func calculateBlockHash(blockNumber int64, timestamp, previousHash string, events []string) string {
	// Create a data structure for hashing (excluding the hash itself)
	hashData := map[string]interface{}{
		"blockNumber":  blockNumber,
		"timestamp":    timestamp,
		"previousHash": previousHash,
		"events":       events,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(hashData)
	if err != nil {
		return ""
	}

	// Calculate SHA256 hash
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

func (h *Handler) getNextBlockNumber() int64 {
	previousBlockNum := h.getPreviousBlockNumber()
	if previousBlockNum == 0 {
		return 1 // First block
	}
	return previousBlockNum + 1
}

// submitToOrderer submits a transaction to the orderer cluster via gRPC
func (h *Handler) submitToOrderer(tx *proto.Transaction) error {
	if h.ordererClient == nil {
		fmt.Printf("Orderer client not initialized, skipping transaction submission\n")
		return nil // Don't fail if orderer is not available
	}

	return h.ordererClient.SubmitTransaction(h.peerID, tx)
}

// submitEventToOrderer submits an event to the orderer cluster via gRPC
func (h *Handler) submitEventToOrderer(event *proto.Event) error {
	if h.ordererClient == nil {
		fmt.Printf("Orderer client not initialized, skipping event submission\n")
		return nil // Don't fail if orderer is not available
	}

	return h.ordererClient.SubmitEvent(h.peerID, event)
}

func NewHandler(db *mongo.Database) *Handler {
	// Initialize gRPC client to orderer
	ordererAddr := os.Getenv("ORDERER_ADDR")
	if ordererAddr == "" {
		ordererAddr = "orderer-ord1:7050"
	}

	peerID := os.Getenv("PEER_NODE_ID")
	if peerID == "" {
		peerID = "peer-anchor"
	}

	var ordererClient *grpcclient.OrdererClient
	client, err := grpcclient.NewOrdererClient(ordererAddr)
	if err != nil {
		fmt.Printf("Failed to create orderer client: %v\n", err)
		// Continue without client for now
	} else {
		ordererClient = client
	}

	return &Handler{
		db:            db,
		ordererClient: ordererClient,
		peerID:        peerID,
	}
}

// CreateContract creates a new contract (Anchor) - New Contract-Token Implementation
func (h *Handler) CreateContract(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID          string        `json:"id"`
		Description string        `json:"description"`
		AnchorId    string        `json:"anchorId"`
		SupplierId  string        `json:"supplierId"` // Primary supplier for token transfer
		BankId      string        `json:"bankId"`
		Amount      float64       `json:"amount"`
		Suppliers   []SupplierDTO `json:"suppliers"` // All suppliers with details
		Approvers   []string      `json:"approvers"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate unique contract ID if not provided
	if req.ID == "" {
		bytes := make([]byte, 16)
		rand.Read(bytes)
		req.ID = hex.EncodeToString(bytes)
	}

	// Create contract with suppliers array
	bankApproved := false // Default to false for new contracts
	contract := map[string]interface{}{
		"_id":          req.ID,
		"description":  req.Description,
		"anchorId":     req.AnchorId,
		"supplierId":   req.SupplierId, // Primary supplier for token transfer
		"bankId":       req.BankId,
		"bankApproved": bankApproved,
		"amount":       req.Amount,
		"suppliers":    req.Suppliers, // All suppliers with details
		"approvers":    req.Approvers,
		"approved":     false,
		"createdAt":    time.Now().Format(time.RFC3339),
	}

	// Save contract to MongoDB
	fmt.Printf("DEBUG: Attempting to save contract with ID: %s\n", contract["_id"])
	_, err := h.db.Collection("contracts").InsertOne(context.Background(), contract)
	if err != nil {
		fmt.Printf("DEBUG: Error saving contract: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("DEBUG: Contract saved successfully\n")

	// Note: Token will be created only after bank approval

	// Log event
	eventId := generateEventID()
	event := map[string]interface{}{
		"eventId":     eventId,
		"eventType":   "CONTRACT_CREATED",
		"contractId":  req.ID,
		"anchorId":    req.AnchorId,
		"bankId":      req.BankId,
		"totalAmount": req.Amount,
		"description": req.Description,
		"suppliers":   req.Suppliers,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	_, err = h.db.Collection("events").InsertOne(context.Background(), event)
	if err != nil {
		fmt.Printf("Failed to log event: %v\n", err)
		// Don't fail the contract creation for logging errors
	}

	// Submit event to public blockchain
	now := time.Now()
	protoTimestamp := timestamppb.New(now)
	protoEvent := &proto.Event{
		EventId:     eventId,
		EventType:   "CONTRACT_CREATED",
		ContractId:  req.ID,
		BankId:      req.BankId,
		Amount:      req.Amount,
		Description: req.Description,
		Timestamp:   protoTimestamp,
	}

	if err := h.submitEventToOrderer(protoEvent); err != nil {
		fmt.Printf("Failed to submit event to orderer: %v\n", err)
		// Continue, don't fail the contract creation
	}

	// Create block entry
	blockNumber := h.getNextBlockNumber()
	timestamp := time.Now().Format(time.RFC3339)
	eventIds := []string{eventId}
	previousHash := h.getPreviousBlockHash()
	blockHash := calculateBlockHash(blockNumber, timestamp, previousHash, eventIds)

	block := map[string]interface{}{
		"blockNumber":  blockNumber,
		"timestamp":    timestamp,
		"events":       eventIds,
		"previousHash": previousHash,
		"hash":         blockHash,
	}

	_, err = h.db.Collection("blocks").InsertOne(context.Background(), block)
	if err != nil {
		// Log error but don't fail the contract creation
		fmt.Printf("Failed to create block: %v\n", err)
	} else {
		// Submit transaction to orderer via gRPC
		now := time.Now()
		protoTimestamp := timestamppb.New(now)
		tx := &proto.Transaction{
			TransactionId:   generateEventID(),
			TransactionType: "CONTRACT_CREATED",
			ContractId:      req.ID,
			SenderId:        req.AnchorId,
			ReceiverId:      req.BankId,
			Amount:          req.Amount,
			Payload:         fmt.Sprintf("Contract created: %s", req.Description),
			Timestamp:       protoTimestamp,
		}

		if err := h.submitToOrderer(tx); err != nil {
			fmt.Printf("Failed to submit transaction to orderer: %v\n", err)
			// Continue, don't fail the contract creation
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"contractId": req.ID,
		"status":     "success",
	})
}

// ApproveContract allows suppliers to approve contracts
func (h *Handler) ApproveContract(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contractId := vars["id"]

	var req struct {
		SupplierId string `json:"supplierId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("DEBUG: Approving contract %s for supplier %s\n", contractId, req.SupplierId)

	// Get contract to verify bank approval and supplier authorization
	var contract map[string]interface{}
	err := h.db.Collection("contracts").FindOne(context.Background(), bson.M{"_id": contractId}).Decode(&contract)
	if err != nil {
		http.Error(w, "Contract not found", http.StatusNotFound)
		return
	}

	// Verify contract has been approved by bank
	bankApproved, ok := contract["bankApproved"].(bool)
	if !ok || !bankApproved {
		http.Error(w, "Contract must be approved by bank before supplier approval", http.StatusForbidden)
		return
	}

	// Verify supplier is part of this contract
	isValidSupplier := false
	if suppliersArray, ok := contract["suppliers"].(primitive.A); ok {
		for _, supplierInterface := range suppliersArray {
			var supplier map[string]interface{}
			if s, ok := supplierInterface.(map[string]interface{}); ok {
				supplier = s
			} else if s, ok := supplierInterface.(primitive.M); ok {
				supplier = map[string]interface{}(s)
			} else {
				continue
			}

			if supplierId, ok := supplier["supplierId"].(string); ok && supplierId == req.SupplierId {
				isValidSupplier = true
				break
			} else if supplierId, ok := supplier["supplierid"].(string); ok && supplierId == req.SupplierId {
				// Try lowercase variant
				isValidSupplier = true
				break
			} else if supplier["supplierId"] == nil {
				// Check if supplierId is null and supplier name matches user username
				if supplierName, ok := supplier["name"].(string); ok {
					// Get user by supplierId to get their username
					var user map[string]interface{}
					err := h.db.Collection("users").FindOne(context.Background(), bson.M{"id": req.SupplierId}).Decode(&user)
					if err == nil {
						if username, ok := user["username"].(string); ok && username == supplierName {
							isValidSupplier = true
							break
						}
					}
				}
			}
		}
	} else {
		http.Error(w, "Invalid contract suppliers data", http.StatusInternalServerError)
		return
	}

	if !isValidSupplier {
		http.Error(w, "Supplier is not authorized to approve this contract", http.StatusForbidden)
		return
	}

	// Define filter for contract updates
	filter := bson.M{"_id": contractId}

	// Update individual supplier status
	if suppliersArray, ok := contract["suppliers"].(primitive.A); ok {
		for i, supplierInterface := range suppliersArray {
			var supplier map[string]interface{}
			if s, ok := supplierInterface.(map[string]interface{}); ok {
				supplier = s
			} else if s, ok := supplierInterface.(primitive.M); ok {
				supplier = map[string]interface{}(s)
			} else {
				continue
			}

			// Check if this is the approving supplier
			isApprovingSupplier := false
			if supplierId, ok := supplier["supplierId"].(string); ok && supplierId == req.SupplierId {
				isApprovingSupplier = true
			} else if supplier["supplierId"] == nil {
				// Check if supplierId is null and supplier name matches user username
				if supplierName, ok := supplier["name"].(string); ok {
					var user map[string]interface{}
					err := h.db.Collection("users").FindOne(context.Background(), bson.M{"id": req.SupplierId}).Decode(&user)
					if err == nil {
						if username, ok := user["username"].(string); ok && username == supplierName {
							isApprovingSupplier = true
						}
					}
				}
			}

			if isApprovingSupplier {
				// Update this supplier's status to APPROVED
				// First, get the current suppliers array and update it
				if suppliersArray, ok := contract["suppliers"].(primitive.A); ok {
					// Create updated suppliers array
					var updatedSuppliers []map[string]interface{}
					for j, supInterface := range suppliersArray {
						var sup map[string]interface{}
						if s, ok := supInterface.(map[string]interface{}); ok {
							sup = make(map[string]interface{})
							for k, v := range s {
								sup[k] = v
							}
						} else if s, ok := supInterface.(primitive.M); ok {
							sup = make(map[string]interface{})
							for k, v := range s {
								sup[k] = v
							}
						}
						if j == i {
							sup["status"] = "APPROVED"
						}
						updatedSuppliers = append(updatedSuppliers, sup)
					}

					// Update the entire suppliers array
					statusUpdate := bson.M{"$set": bson.M{"suppliers": updatedSuppliers}}
					_, err = h.db.Collection("contracts").UpdateOne(context.Background(), filter, statusUpdate)
					if err != nil {
						// Don't return error here, continue with contract approval
					}
				}
				break
			}
		}
	}

	// Refetch contract data to get updated supplier statuses
	var updatedContract map[string]interface{}
	err = h.db.Collection("contracts").FindOne(context.Background(), bson.M{"_id": contractId}).Decode(&updatedContract)
	if err != nil {
		fmt.Printf("DEBUG: Error refetching contract: %v\n", err)
		// Continue with old contract data
		updatedContract = contract
	}

	// Check if all suppliers have approved using updated data
	allApproved := true
	if suppliersArray, ok := updatedContract["suppliers"].(primitive.A); ok {
		for _, supplierInterface := range suppliersArray {
			var supplier map[string]interface{}
			if s, ok := supplierInterface.(map[string]interface{}); ok {
				supplier = s
			} else if s, ok := supplierInterface.(primitive.M); ok {
				supplier = map[string]interface{}(s)
			} else {
				continue
			}

			status, ok := supplier["status"].(string)
			if !ok || status != "APPROVED" {
				allApproved = false
				break
			}
		}
	}

	tokenId := fmt.Sprintf("token_%s", contractId)

	if allApproved {
		// All suppliers approved - distribute tokens proportionally

		// Update contract approval status
		update := bson.M{"$set": bson.M{"approved": true}}
		_, err = h.db.Collection("contracts").UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Distribute token balances to all approved suppliers
		if suppliersArray, ok := contract["suppliers"].(primitive.A); ok {
			for _, supplierInterface := range suppliersArray {
				var supplier map[string]interface{}
				if s, ok := supplierInterface.(map[string]interface{}); ok {
					supplier = s
				} else if s, ok := supplierInterface.(primitive.M); ok {
					supplier = map[string]interface{}(s)
				} else {
					continue
				}

				// Get supplier ID
				supplierId := ""
				if id, ok := supplier["supplierId"].(string); ok && id != "" {
					supplierId = id
				} else if name, ok := supplier["name"].(string); ok {
					// Find user by name to get ID
					var user map[string]interface{}
					err := h.db.Collection("users").FindOne(context.Background(), bson.M{"username": name}).Decode(&user)
					if err == nil {
						if id, ok := user["id"].(string); ok {
							supplierId = id
						}
					}
				}

				if supplierId != "" {
					// Get supplier amount
					amount := 0.0
					if amt, ok := supplier["amount"].(float64); ok {
						amount = amt
					}

					// Create balance for this supplier
					balance := models.Balance{
						TokenId: tokenId,
						Account: supplierId,
						Balance: amount,
					}

					_, err = h.db.Collection("balances").InsertOne(context.Background(), balance)
					if err != nil {
						// Continue with other suppliers
					}
				}
			}
		}

		// Remove anchor's balance since tokens are now distributed
		h.db.Collection("balances").DeleteOne(context.Background(), bson.M{
			"tokenId": tokenId,
			"account": "ANCHOR001",
		})
	}

	// Log approval event
	eventType := "SUPPLIER_APPROVED"
	if allApproved {
		eventType = "CONTRACT_FULLY_APPROVED"
	}

	eventId := generateEventID()
	event := map[string]interface{}{
		"eventId":    eventId,
		"eventType":  eventType,
		"contractId": contractId,
		"tokenId":    tokenId,
		"supplierId": req.SupplierId,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	_, err = h.db.Collection("events").InsertOne(context.Background(), event)
	if err != nil {
		fmt.Printf("DEBUG: Error logging event: %v\n", err)
		// Don't fail the approval for logging errors
	}

	// Submit event to public blockchain
	now := time.Now()
	protoTimestamp := timestamppb.New(now)
	protoEvent := &proto.Event{
		EventId:     eventId,
		EventType:   eventType,
		ContractId:  contractId,
		TokenId:     tokenId,
		SupplierId:  req.SupplierId,
		Amount:      0,
		Description: fmt.Sprintf("Contract %s approved by supplier %s", contractId, req.SupplierId),
		Timestamp:   protoTimestamp,
	}

	if err := h.submitEventToOrderer(protoEvent); err != nil {
		fmt.Printf("Failed to submit event to orderer: %v\n", err)
		// Continue, don't fail the approval
	}

	// Create block entry
	blockNumber := h.getNextBlockNumber()
	timestamp := time.Now().Format(time.RFC3339)
	eventIds := []string{eventId}
	previousHash := h.getPreviousBlockHash()
	blockHash := calculateBlockHash(blockNumber, timestamp, previousHash, eventIds)

	block := map[string]interface{}{
		"blockNumber":  blockNumber,
		"timestamp":    timestamp,
		"events":       eventIds,
		"previousHash": previousHash,
		"hash":         blockHash,
	}

	_, err = h.db.Collection("blocks").InsertOne(context.Background(), block)
	if err != nil {
		fmt.Printf("DEBUG: Error creating approval block: %v\n", err)
		// Don't fail the approval for block creation errors
	} else {
		// Submit transaction to orderer for public blockchain sync
		now := time.Now()
		protoTimestamp := timestamppb.New(now)
		tx := &proto.Transaction{
			TransactionId:   generateEventID(),
			TransactionType: eventType,
			ContractId:      contractId,
			TokenId:         tokenId,
			SenderId:        req.SupplierId,
			ReceiverId:      "SYSTEM",
			Amount:          0, // Approval transaction, no amount
			Payload:         fmt.Sprintf("Contract %s approved by supplier %s", contractId, req.SupplierId),
			Timestamp:       protoTimestamp,
		}

		if err := h.submitToOrderer(tx); err != nil {
			fmt.Printf("Failed to submit supplier approval to orderer: %v\n", err)
			// Continue, don't fail the approval
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Contract approved successfully",
	})
}

// ApproveContractByBank allows banks to approve contracts
func (h *Handler) ApproveContractByBank(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contractId := vars["id"]

	var req struct {
		BankId string `json:"bankId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("DEBUG: Bank %s approving contract %s\n", req.BankId, contractId)

	// Get contract to verify bank permission
	var contract map[string]interface{}
	err := h.db.Collection("contracts").FindOne(context.Background(), bson.M{"_id": contractId}).Decode(&contract)
	if err != nil {
		http.Error(w, "Contract not found", http.StatusNotFound)
		return
	}

	// Verify bank has permission to approve this contract
	contractBankId, ok := contract["bankId"].(string)
	if !ok || contractBankId != req.BankId {
		http.Error(w, "Bank does not have permission to approve this contract", http.StatusForbidden)
		return
	}

	// Update contract bank approval status
	filter := bson.M{"_id": contractId}
	update := bson.M{"$set": bson.M{"bankApproved": true}}
	_, err = h.db.Collection("contracts").UpdateOne(context.Background(), filter, update)
	if err != nil {
		fmt.Printf("DEBUG: Error updating contract bank approval: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate total amount from all suppliers for token creation
	fmt.Printf("DEBUG: Contract suppliers type: %T, value: %+v\n", contract["suppliers"], contract["suppliers"])

	totalAmount := 0.0

	// Handle the suppliers array - MongoDB driver returns primitive.A
	if suppliersArray, ok := contract["suppliers"].(primitive.A); ok {
		fmt.Printf("DEBUG: Suppliers is primitive.A with %d items\n", len(suppliersArray))
		fmt.Printf("DEBUG: Processing suppliers array with %d items\n", len(suppliersArray))
		for i, supplierInterface := range suppliersArray {
			fmt.Printf("DEBUG: Processing supplier %d: %T\n", i, supplierInterface)

			// Try to convert to map
			var supplier map[string]interface{}
			if s, ok := supplierInterface.(map[string]interface{}); ok {
				supplier = s
			} else if s, ok := supplierInterface.(primitive.M); ok {
				supplier = map[string]interface{}(s)
			} else if s, ok := supplierInterface.(bson.M); ok {
				supplier = map[string]interface{}(s)
			} else {
				fmt.Printf("DEBUG: Cannot convert supplier %d to map\n", i)
				continue
			}

			if amount, ok := supplier["amount"].(float64); ok {
				fmt.Printf("DEBUG: Adding amount %f from supplier %d\n", amount, i)
				totalAmount += amount
			} else {
				fmt.Printf("DEBUG: No valid amount found in supplier %d\n", i)
			}
		}
	} else {
		fmt.Printf("DEBUG: Suppliers is not []interface{}, type: %T\n", contract["suppliers"])
		http.Error(w, "Invalid contract suppliers data", http.StatusInternalServerError)
		return
	}

	// Create token after bank approval
	tokenId := fmt.Sprintf("token_%s", contractId)
	symbol := fmt.Sprintf("TK%s", contractId[len(contractId)-4:]) // Last 4 chars of contract ID

	// Token should be owned by anchor (buyer) after bank approval
	anchorId := "ANCHOR001" // Default anchor ID

	token := models.Token{
		ID:         tokenId,
		ContractId: contractId,
		Symbol:     symbol,
		Total:      totalAmount,
		Issuer:     "SYSTEM", // System auto-generates token for anchor
		Owner:      anchorId, // Owned by anchor after bank approval
		CreatedAt:  time.Now().Format(time.RFC3339),
	}

	_, err = h.db.Collection("tokens").InsertOne(context.Background(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create initial balance for anchor (token owner)
	balance := models.Balance{
		TokenId: tokenId,
		Account: anchorId,
		Balance: totalAmount,
	}

	_, err = h.db.Collection("balances").InsertOne(context.Background(), balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log bank approval and token generation event
	eventId := generateEventID()
	event := map[string]interface{}{
		"eventId":     eventId,
		"eventType":   "CONTRACT_BANK_APPROVED_TOKEN_GENERATED",
		"contractId":  contractId,
		"tokenId":     tokenId,
		"bankId":      req.BankId,
		"anchorId":    anchorId,
		"totalAmount": totalAmount,
		"description": "Bank approved contract and system auto-generated token for anchor",
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	_, err = h.db.Collection("events").InsertOne(context.Background(), event)
	if err != nil {
		fmt.Printf("DEBUG: Error logging bank approval event: %v\n", err)
		// Don't fail the approval for logging errors
	}

	// Submit event to public blockchain
	now := time.Now()
	protoTimestamp := timestamppb.New(now)
	protoEvent := &proto.Event{
		EventId:     eventId,
		EventType:   "CONTRACT_BANK_APPROVED_TOKEN_GENERATED",
		ContractId:  contractId,
		TokenId:     tokenId,
		BankId:      req.BankId,
		Amount:      totalAmount,
		Description: "Bank approved contract and system auto-generated token for anchor",
		Timestamp:   protoTimestamp,
	}

	if err := h.submitEventToOrderer(protoEvent); err != nil {
		fmt.Printf("Failed to submit event to orderer: %v\n", err)
		// Continue, don't fail the approval
	}

	// Create block entry
	blockNumber := h.getNextBlockNumber()
	timestamp := time.Now().Format(time.RFC3339)
	eventIds := []string{eventId}
	previousHash := h.getPreviousBlockHash()
	blockHash := calculateBlockHash(blockNumber, timestamp, previousHash, eventIds)

	block := map[string]interface{}{
		"blockNumber":  blockNumber,
		"timestamp":    timestamp,
		"events":       eventIds,
		"previousHash": previousHash,
		"hash":         blockHash,
	}

	_, err = h.db.Collection("blocks").InsertOne(context.Background(), block)
	if err != nil {
		fmt.Printf("DEBUG: Error creating bank approval block: %v\n", err)
		// Don't fail the approval for block creation errors
	} else {
		// Submit transaction to orderer for public blockchain sync
		now := time.Now()
		protoTimestamp := timestamppb.New(now)
		tx := &proto.Transaction{
			TransactionId:   generateEventID(),
			TransactionType: "CONTRACT_BANK_APPROVED_TOKEN_GENERATED",
			ContractId:      contractId,
			TokenId:         tokenId,
			SenderId:        req.BankId,
			ReceiverId:      "ANCHOR001",
			Amount:          totalAmount,
			Payload:         fmt.Sprintf("Contract %s approved by bank %s, token %s generated", contractId, req.BankId, tokenId),
			Timestamp:       protoTimestamp,
		}

		if err := h.submitToOrderer(tx); err != nil {
			fmt.Printf("Failed to submit bank approval to orderer: %v\n", err)
			// Continue, don't fail the approval
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Contract approved by bank successfully",
		"tokenId": tokenId,
	})
}

// GetContracts returns all contracts
func (h *Handler) GetContracts(w http.ResponseWriter, r *http.Request) {
	cursor, err := h.db.Collection("contracts").Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var contracts []map[string]interface{}
	if err = cursor.All(context.Background(), &contracts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contracts)
}

// GetContract returns contract details
func (h *Handler) GetContract(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contractId := vars["id"]

	var contract map[string]interface{}
	err := h.db.Collection("contracts").FindOne(context.Background(), bson.M{"_id": contractId}).Decode(&contract)
	if err != nil {
		http.Error(w, "Contract not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contract)
}

// GetToken returns token information
func (h *Handler) GetToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenId := vars["id"]

	var token models.Token
	err := h.db.Collection("tokens").FindOne(context.Background(), bson.M{"_id": tokenId}).Decode(&token)
	if err != nil {
		http.Error(w, "Token not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

// TransferToken allows suppliers to transfer tokens
func (h *Handler) TransferToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TokenId string  `json:"tokenId"`
		From    string  `json:"from"`
		To      string  `json:"to"`
		Amount  float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get token info
	var token models.Token
	err := h.db.Collection("tokens").FindOne(context.Background(), bson.M{"_id": req.TokenId}).Decode(&token)
	if err != nil {
		http.Error(w, "Token not found", http.StatusNotFound)
		return
	}

	// Check if sender has balance for this token (not necessarily the owner)
	// Allow anyone with balance to transfer their portion

	// Get sender balance
	var fromBalance models.Balance
	err = h.db.Collection("balances").FindOne(
		context.Background(),
		bson.M{"tokenId": req.TokenId, "account": req.From},
	).Decode(&fromBalance)
	if err != nil {
		http.Error(w, "Sender has no balance", http.StatusBadRequest)
		return
	}

	if fromBalance.Balance < req.Amount {
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}

	// Update sender balance
	_, err = h.db.Collection("balances").UpdateOne(
		context.Background(),
		bson.M{"tokenId": req.TokenId, "account": req.From},
		bson.M{"$set": bson.M{"balance": fromBalance.Balance - req.Amount}},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update receiver balance
	var toBalance models.Balance
	err = h.db.Collection("balances").FindOne(
		context.Background(),
		bson.M{"tokenId": req.TokenId, "account": req.To},
	).Decode(&toBalance)

	if err == mongo.ErrNoDocuments {
		// Create new balance entry for receiver
		receiverBalance := models.Balance{
			TokenId:         req.TokenId,
			Account:         req.To,
			Balance:         req.Amount,
			TransferredFrom: req.From,
		}
		_, err = h.db.Collection("balances").InsertOne(context.Background(), receiverBalance)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		// Update existing balance
		_, err = h.db.Collection("balances").UpdateOne(
			context.Background(),
			bson.M{"tokenId": req.TokenId, "account": req.To},
			bson.M{"$set": bson.M{
				"balance":         toBalance.Balance + req.Amount,
				"transferredFrom": req.From,
			}},
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Log transfer event
	eventId := generateEventID()
	event := map[string]interface{}{
		"eventId":   eventId,
		"eventType": "TOKEN_TRANSFERRED",
		"tokenId":   req.TokenId,
		"from":      req.From,
		"to":        req.To,
		"amount":    req.Amount,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	_, err = h.db.Collection("events").InsertOne(context.Background(), event)
	if err != nil {
		// Log error but don't fail the transfer
		fmt.Printf("Failed to log transfer event: %v\n", err)
	}

	// Submit event to public blockchain
	now := time.Now()
	protoTimestamp := timestamppb.New(now)
	protoEvent := &proto.Event{
		EventId:     eventId,
		EventType:   "TOKEN_TRANSFERRED",
		TokenId:     req.TokenId,
		From:        req.From,
		To:          req.To,
		Amount:      req.Amount,
		Description: fmt.Sprintf("Token transfer: %s -> %s, amount: %.2f", req.From, req.To, req.Amount),
		Timestamp:   protoTimestamp,
	}

	if err := h.submitEventToOrderer(protoEvent); err != nil {
		fmt.Printf("Failed to submit event to orderer: %v\n", err)
		// Continue, don't fail the transfer
	}

	// Create block entry
	blockNumber := h.getNextBlockNumber()
	timestamp := time.Now().Format(time.RFC3339)
	eventIds := []string{eventId}
	previousHash := h.getPreviousBlockHash()
	blockHash := calculateBlockHash(blockNumber, timestamp, previousHash, eventIds)

	block := map[string]interface{}{
		"blockNumber":  blockNumber,
		"timestamp":    timestamp,
		"events":       eventIds,
		"previousHash": previousHash,
		"hash":         blockHash,
	}

	_, err = h.db.Collection("blocks").InsertOne(context.Background(), block)
	if err != nil {
		// Log error but don't fail the transfer
		fmt.Printf("Failed to create transfer block: %v\n", err)
	} else {
		// Submit transaction to orderer for public blockchain sync
		now := time.Now()
		protoTimestamp := timestamppb.New(now)
		tx := &proto.Transaction{
			TransactionId:   generateEventID(),
			TransactionType: "TOKEN_TRANSFERRED",
			ContractId:      strings.TrimPrefix(req.TokenId, "token_"), // Extract contract ID
			TokenId:         req.TokenId,
			SenderId:        req.From,
			ReceiverId:      req.To,
			Amount:          req.Amount,
			Payload:         fmt.Sprintf("Token transfer: %s -> %s, amount: %.2f", req.From, req.To, req.Amount),
			Timestamp:       protoTimestamp,
		}

		if err := h.submitToOrderer(tx); err != nil {
			fmt.Printf("Failed to submit token transfer to orderer: %v\n", err)
			// Continue, don't fail the transfer
		}
	}

	// Check if anchor has no more balance for this token
	// If so, automatically approve the contract
	anchorBalanceFilter := bson.M{"tokenId": req.TokenId, "account": "ANCHOR001"}
	var anchorBalance models.Balance
	err = h.db.Collection("balances").FindOne(context.Background(), anchorBalanceFilter).Decode(&anchorBalance)

	// If anchor has no balance or balance is 0, approve the contract
	if err == mongo.ErrNoDocuments || anchorBalance.Balance <= 0 {
		// Extract contract ID from token ID (format: token_{contractId})
		contractId := strings.TrimPrefix(req.TokenId, "token_")

		// Update contract status to approved
		contractFilter := bson.M{"_id": contractId}
		contractUpdate := bson.M{"$set": bson.M{"approved": true, "status": "APPROVED"}}
		_, err = h.db.Collection("contracts").UpdateOne(context.Background(), contractFilter, contractUpdate)
		if err != nil {
			fmt.Printf("Failed to auto-approve contract %s: %v\n", contractId, err)
		} else {
			fmt.Printf("Auto-approved contract %s after complete token transfer\n", contractId)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "transferred",
		"message": "Token transferred successfully",
	})
}

// GetTokensIssuedByBank returns all tokens issued by a specific bank
func (h *Handler) GetTokensIssuedByBank(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bankId := vars["bankId"]

	cursor, err := h.db.Collection("tokens").Find(context.Background(), bson.M{"issuer": bankId})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var tokens []models.Token
	if err = cursor.All(context.Background(), &tokens); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

// GetContractLedger returns all events and blocks related to a contract
func (h *Handler) GetContractLedger(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contractId := vars["id"]

	// Get all events related to this contract
	eventCursor, err := h.db.Collection("events").Find(context.Background(), bson.M{"contractId": contractId})
	if err != nil {
		http.Error(w, "Failed to query events: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer eventCursor.Close(context.Background())

	var events []map[string]interface{}
	if err = eventCursor.All(context.Background(), &events); err != nil {
		http.Error(w, "Failed to decode events: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get all events related to this contract's token (token_{contractId})
	tokenId := "token_" + contractId
	tokenEventCursor, err := h.db.Collection("events").Find(context.Background(), bson.M{"tokenId": tokenId})
	if err != nil {
		http.Error(w, "Failed to query token events: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer tokenEventCursor.Close(context.Background())

	var tokenEvents []map[string]interface{}
	if err = tokenEventCursor.All(context.Background(), &tokenEvents); err != nil {
		http.Error(w, "Failed to decode token events: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Combine all events
	allEvents := append(events, tokenEvents...)

	// Get all blocks containing these events
	var eventIds []string
	for _, event := range allEvents {
		if eventId, ok := event["eventId"].(string); ok {
			eventIds = append(eventIds, eventId)
		}
	}

	blockCursor, err := h.db.Collection("blocks").Find(context.Background(), bson.M{"events": bson.M{"$in": eventIds}})
	if err != nil {
		http.Error(w, "Failed to query blocks: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer blockCursor.Close(context.Background())

	var blocks []map[string]interface{}
	if err = blockCursor.All(context.Background(), &blocks); err != nil {
		http.Error(w, "Failed to decode blocks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get token balances for this contract
	balanceCursor, err := h.db.Collection("balances").Find(context.Background(), bson.M{"tokenId": tokenId})
	if err != nil {
		http.Error(w, "Failed to query balances: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer balanceCursor.Close(context.Background())

	var balances []map[string]interface{}
	if err = balanceCursor.All(context.Background(), &balances); err != nil {
		http.Error(w, "Failed to decode balances: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare ledger response
	ledger := map[string]interface{}{
		"contractId": contractId,
		"tokenId":    tokenId,
		"events":     allEvents,
		"blocks":     blocks,
		"balances":   balances,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ledger)
}

// UpdateBlockHashes recalculates and updates hashes for all blocks
func (h *Handler) UpdateBlockHashes(w http.ResponseWriter, r *http.Request) {
	// Get all blocks sorted by block number
	cursor, err := h.db.Collection("blocks").Find(context.Background(), bson.M{}, options.Find().SetSort(bson.M{"blockNumber": 1}))
	if err != nil {
		http.Error(w, "Failed to query blocks: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var blocks []map[string]interface{}
	if err = cursor.All(context.Background(), &blocks); err != nil {
		http.Error(w, "Failed to decode blocks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	updatedCount := 0
	for _, block := range blocks {
		// Get block data
		blockNumber := int64(0)
		if bn, ok := block["blockNumber"].(int64); ok {
			blockNumber = bn
		} else if bn, ok := block["blockNumber"].(int32); ok {
			blockNumber = int64(bn)
		}

		timestamp := ""
		if ts, ok := block["timestamp"].(string); ok {
			timestamp = ts
		}

		var eventIds []string
		if events, ok := block["events"].([]interface{}); ok {
			for _, event := range events {
				if eventId, ok := event.(string); ok {
					eventIds = append(eventIds, eventId)
				}
			}
		}

		// Determine previous hash - genesis for first block, otherwise get from previous block
		previousHash := "genesis"
		if blockNumber > 1 {
			// Get previous block's hash from database
			prevBlockFilter := bson.M{"blockNumber": blockNumber - 1}
			var prevBlock map[string]interface{}
			err := h.db.Collection("blocks").FindOne(context.Background(), prevBlockFilter).Decode(&prevBlock)
			if err == nil {
				if ph, ok := prevBlock["hash"].(string); ok && ph != "" {
					previousHash = ph
				}
			}
		}

		// Calculate new hash
		newHash := calculateBlockHash(blockNumber, timestamp, previousHash, eventIds)

		// Update block in database
		filter := bson.M{"_id": block["_id"]}
		update := bson.M{"$set": bson.M{
			"hash":         newHash,
			"previousHash": previousHash,
		}}

		_, err = h.db.Collection("blocks").UpdateOne(context.Background(), filter, update)
		if err != nil {
			fmt.Printf("Failed to update block %d: %v\n", blockNumber, err)
			continue
		}

		updatedCount++
		fmt.Printf("Updated block %d with hash: %s, previous: %s\n", blockNumber, newHash, previousHash)
	}

	response := map[string]interface{}{
		"message":      "Block hashes updated successfully",
		"updatedCount": updatedCount,
		"totalBlocks":  len(blocks),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAllTokens returns all tokens
func (h *Handler) GetAllTokens(w http.ResponseWriter, r *http.Request) {
	cursor, err := h.db.Collection("tokens").Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var tokens []models.Token
	if err = cursor.All(context.Background(), &tokens); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

// GetBalancesByAccount returns all balances for a specific account
func (h *Handler) GetBalancesByAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountId := vars["accountId"]

	cursor, err := h.db.Collection("balances").Find(context.Background(), bson.M{"account": accountId})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var balances []map[string]interface{}
	if err = cursor.All(context.Background(), &balances); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balances)
}

// GetBalancesByToken returns all balances for a specific token
func (h *Handler) GetBalancesByToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenId := vars["tokenId"]

	cursor, err := h.db.Collection("balances").Find(context.Background(), bson.M{"tokenId": tokenId})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var balances []map[string]interface{}
	if err = cursor.All(context.Background(), &balances); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balances)
}

// SettleToken allows suppliers to settle their tokens with the bank, removing their token balance
func (h *Handler) SettleToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TokenId    string `json:"tokenId"`
		SupplierId string `json:"supplierId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("DEBUG: Supplier %s settling token %s\n", req.SupplierId, req.TokenId)

	// Verify supplier has balance for this token
	var balance models.Balance
	err := h.db.Collection("balances").FindOne(
		context.Background(),
		bson.M{"tokenId": req.TokenId, "account": req.SupplierId},
	).Decode(&balance)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Supplier has no balance for this token", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get token info to verify it exists and get bank info
	var token models.Token
	err = h.db.Collection("tokens").FindOne(context.Background(), bson.M{"_id": req.TokenId}).Decode(&token)
	if err != nil {
		http.Error(w, "Token not found", http.StatusNotFound)
		return
	}

	// Get contract info to get bank ID
	var contract map[string]interface{}
	err = h.db.Collection("contracts").FindOne(context.Background(), bson.M{"_id": token.ContractId}).Decode(&contract)
	if err != nil {
		http.Error(w, "Contract not found", http.StatusNotFound)
		return
	}

	bankId, ok := contract["bankId"].(string)
	if !ok {
		http.Error(w, "Invalid contract data - missing bank ID", http.StatusInternalServerError)
		return
	}

	// Remove supplier's balance for this token
	_, err = h.db.Collection("balances").DeleteOne(
		context.Background(),
		bson.M{"tokenId": req.TokenId, "account": req.SupplierId},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log settlement event
	eventId := generateEventID()
	event := map[string]interface{}{
		"eventId":       eventId,
		"eventType":     "TOKEN_SETTLED",
		"tokenId":       req.TokenId,
		"contractId":    token.ContractId,
		"supplierId":    req.SupplierId,
		"bankId":        bankId,
		"settledAmount": balance.Balance,
		"description":   fmt.Sprintf("Supplier %s settled %.2f tokens with bank %s", req.SupplierId, balance.Balance, bankId),
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	_, err = h.db.Collection("events").InsertOne(context.Background(), event)
	if err != nil {
		fmt.Printf("DEBUG: Error logging settlement event: %v\n", err)
		// Don't fail the settlement for logging errors
	}

	// Submit event to public blockchain
	now := time.Now()
	protoTimestamp := timestamppb.New(now)
	protoEvent := &proto.Event{
		EventId:     eventId,
		EventType:   "TOKEN_SETTLED",
		ContractId:  token.ContractId,
		TokenId:     req.TokenId,
		SupplierId:  req.SupplierId,
		BankId:      bankId,
		Amount:      balance.Balance,
		Description: fmt.Sprintf("Supplier %s settled %.2f tokens with bank %s", req.SupplierId, balance.Balance, bankId),
		Timestamp:   protoTimestamp,
	}

	if err := h.submitEventToOrderer(protoEvent); err != nil {
		fmt.Printf("Failed to submit event to orderer: %v\n", err)
		// Continue, don't fail the settlement
	}

	// Create block entry
	blockNumber := h.getNextBlockNumber()
	timestamp := time.Now().Format(time.RFC3339)
	eventIds := []string{eventId}
	previousHash := h.getPreviousBlockHash()
	blockHash := calculateBlockHash(blockNumber, timestamp, previousHash, eventIds)

	block := map[string]interface{}{
		"blockNumber":  blockNumber,
		"timestamp":    timestamp,
		"events":       eventIds,
		"previousHash": previousHash,
		"hash":         blockHash,
	}

	_, err = h.db.Collection("blocks").InsertOne(context.Background(), block)
	if err != nil {
		fmt.Printf("DEBUG: Error creating settlement block: %v\n", err)
		// Don't fail the settlement for block creation errors
	} else {
		// Submit transaction to orderer for public blockchain sync
		now := time.Now()
		protoTimestamp := timestamppb.New(now)
		tx := &proto.Transaction{
			TransactionId:   generateEventID(),
			TransactionType: "TOKEN_SETTLED",
			ContractId:      token.ContractId,
			TokenId:         req.TokenId,
			SenderId:        req.SupplierId,
			ReceiverId:      bankId,
			Amount:          balance.Balance,
			Payload:         fmt.Sprintf("Token settlement: supplier %s settled %.2f tokens with bank %s", req.SupplierId, balance.Balance, bankId),
			Timestamp:       protoTimestamp,
		}

		if err := h.submitToOrderer(tx); err != nil {
			fmt.Printf("Failed to submit token settlement to orderer: %v\n", err)
			// Continue, don't fail the settlement
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "settled",
		"message":       "Token settled successfully with bank",
		"tokenId":       req.TokenId,
		"supplierId":    req.SupplierId,
		"bankId":        bankId,
		"settledAmount": balance.Balance,
	})
}

// GetSuppliers returns all users with role SUPPLIER
func (h *Handler) GetSuppliers(w http.ResponseWriter, r *http.Request) {
	cursor, err := h.db.Collection("users").Find(context.Background(), bson.M{"role": "SUPPLIER"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var suppliers []map[string]interface{}
	if err = cursor.All(context.Background(), &suppliers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suppliers)
}
