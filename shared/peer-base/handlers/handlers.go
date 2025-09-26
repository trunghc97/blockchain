package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	timestamp_pb "google.golang.org/protobuf/types/known/timestamppb"

	"shared/peer-base/models"
	pb "shared/proto"
)

type PeerNode interface {
	SubmitToOrderer(block *pb.Block) error
	PublishEventToKafka(event map[string]interface{}, channel string) error
	PublishTransactionToKafka(transaction *pb.Transaction, channel string) error
}

type Handler struct {
	db       *mongo.Database
	nodeType string
	peerNode PeerNode
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

func generateEventID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func NewHandler(db *mongo.Database, nodeType string, peerNode PeerNode) *Handler {
	return &Handler{
		db:       db,
		nodeType: nodeType,
		peerNode: peerNode,
	}
}

// CreateContract creates a new contract (Anchor) - New Contract-Token Implementation
func (h *Handler) CreateContract(w http.ResponseWriter, r *http.Request) {
	// Only anchor peers can create contracts
	if h.nodeType != "anchor" {
		http.Error(w, "Only anchor peers can create contracts", http.StatusForbidden)
		return
	}

	var req struct {
		ID          string        `json:"id"`
		Description string        `json:"description"`
		AnchorId    string        `json:"anchorId"`
		SupplierId  string        `json:"supplierId"` // Primary supplier for token transfer
		BankId      string        `json:"bankId"`
		Amount      float64       `json:"amount"`
		Suppliers   []interface{} `json:"suppliers"` // All suppliers with details
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

	// Log event
	event := map[string]interface{}{
		"eventId":     generateEventID(),
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
		// Log error but don't fail the contract creation
		fmt.Printf("Failed to log event: %v\n", err)
	}

	// Create block entry
	blockNumber := h.getNextBlockNumber()
	timestamp := time.Now().Format(time.RFC3339)
	eventIds := []string{event["eventId"].(string)}
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
	}

	// Publish event to Kafka instead of calling Orderer directly
	go func() {
		channel := "scf" // Default channel for SCF transactions
		if h.nodeType == "main-bank" && (event["eventType"] == "BANK_APPROVED_TOKEN_GENERATED") {
			channel = "audit" // Bank approvals go to audit channel
		}

		err := h.peerNode.PublishEventToKafka(event, channel)
		if err != nil {
			fmt.Printf("Failed to publish event to Kafka: %v\n", err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"contractId": req.ID,
		"status":     "success",
	})
}

// ApproveContract allows suppliers to approve contracts
func (h *Handler) ApproveContract(w http.ResponseWriter, r *http.Request) {
	// Only supplier peers can approve contracts
	if h.nodeType != "supplier" {
		http.Error(w, "Only supplier peers can approve contracts", http.StatusForbidden)
		return
	}

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
			}
		}
	}

	if !isValidSupplier {
		http.Error(w, "Supplier is not authorized to approve this contract", http.StatusForbidden)
		return
	}

	// Update supplier status and check if all approved
	// ... (rest of the approval logic similar to original)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Contract approved successfully",
	})
}

// ApproveContractByBank allows banks to approve contracts
func (h *Handler) ApproveContractByBank(w http.ResponseWriter, r *http.Request) {
	// Only main-bank peers can approve contracts by bank
	if h.nodeType != "main-bank" {
		http.Error(w, "Only main-bank peers can approve contracts", http.StatusForbidden)
		return
	}

	// ... (bank approval logic)
}

// TransferToken allows suppliers to transfer tokens
func (h *Handler) TransferToken(w http.ResponseWriter, r *http.Request) {
	// Token transfers can happen on any peer type
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

	// ... (token transfer logic)

	// Publish transfer event to Kafka
	go func() {
		transferEvent := map[string]interface{}{
			"eventId":     generateEventID(),
			"eventType":   "TOKEN_TRANSFERRED",
			"tokenId":     req.TokenId,
			"from":        req.From,
			"to":          req.To,
			"amount":      req.Amount,
			"timestamp":   time.Now().Format(time.RFC3339),
		}

		channel := "scf" // Token transfers go to SCF channel
		err := h.peerNode.PublishEventToKafka(transferEvent, channel)
		if err != nil {
			fmt.Printf("Failed to publish transfer event to Kafka: %v\n", err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "transferred",
		"message": "Token transferred successfully",
	})
}

// Peer-specific handler methods
func (h *Handler) GetBankDashboard(w http.ResponseWriter, r *http.Request) {
	// Bank-specific dashboard logic
}

func (h *Handler) GetSupplierDashboard(w http.ResponseWriter, r *http.Request) {
	// Supplier-specific dashboard logic
}

func (h *Handler) GetAnchorDashboard(w http.ResponseWriter, r *http.Request) {
	// Anchor-specific dashboard logic
}

// ... (implement other methods from original handlers.go)

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

		// Determine previous hash
		previousHash := "genesis"
		if blockNumber > 1 {
			// Get previous block's hash
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

		// Update block
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
		fmt.Printf("Updated block %d with hash: %s\n", blockNumber, newHash)
	}

	response := map[string]interface{}{
		"message":      "Block hashes updated successfully",
		"updatedCount": updatedCount,
		"totalBlocks":  len(blocks),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Placeholder implementations for remaining methods
func (h *Handler) GetContract(w http.ResponseWriter, r *http.Request) { /* implement */ }
func (h *Handler) GetContractLedger(w http.ResponseWriter, r *http.Request) { /* implement */ }
func (h *Handler) SettleToken(w http.ResponseWriter, r *http.Request) { /* implement */ }
func (h *Handler) GetTokensIssuedByBank(w http.ResponseWriter, r *http.Request) { /* implement */ }
func (h *Handler) GetBalancesByAccount(w http.ResponseWriter, r *http.Request) { /* implement */ }
func (h *Handler) GetBalancesByToken(w http.ResponseWriter, r *http.Request) { /* implement */ }
func (h *Handler) GetBankIssuedTokens(w http.ResponseWriter, r *http.Request) { /* implement */ }
func (h *Handler) GetBankPendingContracts(w http.ResponseWriter, r *http.Request) { /* implement */ }
func (h *Handler) GetSupplierContracts(w http.ResponseWriter, r *http.Request) { /* implement */ }
func (h *Handler) GetSupplierTokens(w http.ResponseWriter, r *http.Request) { /* implement */ }
func (h *Handler) GetAnchorContracts(w http.ResponseWriter, r *http.Request) { /* implement */ }
func (h *Handler) UploadContractFile(w http.ResponseWriter, r *http.Request) { /* implement */ }