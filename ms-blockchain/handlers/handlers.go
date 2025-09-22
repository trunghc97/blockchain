package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"ms-blockchain/models"
)

type SupplierDTO struct {
	SupplierId string  `json:"supplierId"`
	Name       string  `json:"name"`
	Amount     float64 `json:"amount"`
	Status     string  `json:"status"`
}

type Handler struct {
	db *mongo.Database
}

func generateEventID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (h *Handler) getNextBlockNumber() int64 {
	// Get the highest block number and increment
	var lastBlock map[string]interface{}
	err := h.db.Collection("blocks").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"blockNumber": -1})).Decode(&lastBlock)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 1 // First block
		}
		return 1 // Default to 1 on error
	}

	if blockNum, ok := lastBlock["blockNumber"].(int64); ok {
		return blockNum + 1
	}
	return 1
}

func NewHandler(db *mongo.Database) *Handler {
	return &Handler{db: db}
}

func (h *Handler) generateEventID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
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
	contract := map[string]interface{}{
		"_id":         req.ID,
		"description": req.Description,
		"anchorId":    req.AnchorId,
		"supplierId":  req.SupplierId, // Primary supplier for token transfer
		"bankId":      req.BankId,
		"amount":      req.Amount,
		"suppliers":   req.Suppliers, // All suppliers with details
		"approvers":   req.Approvers,
		"approved":    false,
		"createdAt":   time.Now().Format(time.RFC3339),
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

	// Calculate total amount from all suppliers
	totalAmount := 0.0
	for _, supplier := range req.Suppliers {
		totalAmount += supplier.Amount
	}

	// Auto-issue token when contract is created
	tokenId := fmt.Sprintf("token_%s", req.ID)
	symbol := fmt.Sprintf("TK%s", req.ID[len(req.ID)-4:]) // Last 4 chars of contract ID

	fmt.Printf("DEBUG: Creating token with ID: %s, total: %f\n", tokenId, totalAmount)
	token := models.Token{
		ID:         tokenId,
		ContractId: req.ID,
		Symbol:     symbol,
		Total:      totalAmount,
		Issuer:     req.BankId,
		Owner:      req.AnchorId, // Initially owned by anchor (not bank)
		CreatedAt:  time.Now().Format(time.RFC3339),
	}

	_, err = h.db.Collection("tokens").InsertOne(context.Background(), token)
	if err != nil {
		fmt.Printf("DEBUG: Error saving token: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("DEBUG: Token saved successfully\n")

	// Create initial balance for anchor (not bank)
	balance := models.Balance{
		TokenId: tokenId,
		Account: req.AnchorId,
		Balance: totalAmount,
	}

	_, err = h.db.Collection("balances").InsertOne(context.Background(), balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log event
	event := map[string]interface{}{
		"eventId":     generateEventID(),
		"eventType":   "CONTRACT_CREATED",
		"contractId":  req.ID,
		"tokenId":     tokenId,
		"anchorId":    req.AnchorId,
		"bankId":      req.BankId,
		"totalAmount": totalAmount,
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
	block := map[string]interface{}{
		"blockNumber":  h.getNextBlockNumber(),
		"timestamp":    time.Now().Format(time.RFC3339),
		"events":       []string{event["eventId"].(string)},
		"previousHash": "", // Simplified - in real blockchain this would be hash of previous block
		"hash":         "", // Simplified
	}

	_, err = h.db.Collection("blocks").InsertOne(context.Background(), block)
	if err != nil {
		// Log error but don't fail the contract creation
		fmt.Printf("Failed to create block: %v\n", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"contractId": req.ID,
		"tokenId":    tokenId,
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

	// Update contract approval status
	filter := bson.M{"_id": contractId}
	update := bson.M{"$set": bson.M{"approved": true}}
	_, err := h.db.Collection("contracts").UpdateOne(context.Background(), filter, update)
	if err != nil {
		fmt.Printf("DEBUG: Error updating contract: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Transfer token ownership from anchor to supplier
	tokenId := fmt.Sprintf("token_%s", contractId)
	tokenFilter := bson.M{"_id": tokenId}
	tokenUpdate := bson.M{"$set": bson.M{"owner": req.SupplierId}}
	_, err = h.db.Collection("tokens").UpdateOne(context.Background(), tokenFilter, tokenUpdate)
	if err != nil {
		fmt.Printf("DEBUG: Error updating token owner: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Transfer balance from anchor to supplier
	balanceFilter := bson.M{"tokenId": tokenId, "account": "ANCHOR001"}
	balanceUpdate := bson.M{"$set": bson.M{"account": req.SupplierId}}
	_, err = h.db.Collection("balances").UpdateOne(context.Background(), balanceFilter, balanceUpdate)
	if err != nil {
		fmt.Printf("DEBUG: Error updating balance: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log approval event
	event := map[string]interface{}{
		"eventId":    generateEventID(),
		"eventType":  "CONTRACT_APPROVED",
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "approved",
		"message": "Contract approved successfully",
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

	var contract models.Contract
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

	// Check if sender is the current owner
	if token.Owner != req.From {
		http.Error(w, "Sender is not the current owner", http.StatusForbidden)
		return
	}

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
			TokenId: req.TokenId,
			Account: req.To,
			Balance: req.Amount,
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
			bson.M{"$set": bson.M{"balance": toBalance.Balance + req.Amount}},
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Update token ownership to receiver
	_, err = h.db.Collection("tokens").UpdateOne(
		context.Background(),
		bson.M{"_id": req.TokenId},
		bson.M{"$set": bson.M{"owner": req.To}},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
