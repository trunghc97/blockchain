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
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Handler struct {
	db     *mongo.Database
	peerID string
}

func generateEventID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func NewHandler(db *mongo.Database) *Handler {
	peerID := os.Getenv("PEER_NODE_ID")
	if peerID == "" {
		peerID = "peer-main-bank"
	}

	return &Handler{
		db:     db,
		peerID: peerID,
	}
}

// CreateContract - This method should not exist in peer handlers
// Contract creation should be handled by blockchain-gw via ProposalRequest
func (h *Handler) CreateContract(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "This endpoint is not available. Use blockchain-gw for transaction processing.", http.StatusMethodNotAllowed)
}

// ApproveContract - This method should not exist in peer handlers
// Approval should be handled by blockchain-gw via ProposalRequest
func (h *Handler) ApproveContract(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "This endpoint is not available. Use blockchain-gw for transaction processing.", http.StatusMethodNotAllowed)
}

// ApproveContractByBank - This method should not exist in peer handlers
// Bank approval should be handled by blockchain-gw via ProposalRequest
func (h *Handler) ApproveContractByBank(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "This endpoint is not available. Use blockchain-gw for transaction processing.", http.StatusMethodNotAllowed)
}

// Placeholder methods
func (h *Handler) GetContracts(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *Handler) GetContract(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *Handler) GetContractLedger(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *Handler) GetToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenId := vars["id"]

	// Query token from database
	var token map[string]interface{}
	err := h.db.Collection("tokens").FindOne(context.Background(), bson.M{"_id": tokenId}).Decode(&token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Token not found", http.StatusNotFound)
			return
		}
		fmt.Printf("Error getting token: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

// TransferToken - This method should not exist in peer handlers
// Token transfer should be handled by blockchain-gw via ProposalRequest
func (h *Handler) TransferToken(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "This endpoint is not available. Use blockchain-gw for transaction processing.", http.StatusMethodNotAllowed)
}

// SettleToken - This method should not exist in peer handlers
// Token settlement should be handled by blockchain-gw via ProposalRequest
func (h *Handler) SettleToken(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "This endpoint is not available. Use blockchain-gw for transaction processing.", http.StatusMethodNotAllowed)
}

func (h *Handler) GetTokensIssuedByBank(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bankId := vars["bankId"]

	// Query tokens issued by bank from database
	cursor, err := h.db.Collection("tokens").Find(context.Background(), bson.M{"issuer": bankId})
	if err != nil {
		fmt.Printf("Error getting tokens issued by bank: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var tokens []map[string]interface{}
	for cursor.Next(context.Background()) {
		var token map[string]interface{}
		if err := cursor.Decode(&token); err != nil {
			fmt.Printf("Error decoding token: %v\n", err)
			continue
		}
		tokens = append(tokens, token)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

func (h *Handler) GetSuppliers(w http.ResponseWriter, r *http.Request) {
	// Query suppliers from database (users with role SUPPLIER)
	cursor, err := h.db.Collection("users").Find(context.Background(), bson.M{"role": "SUPPLIER"})
	if err != nil {
		fmt.Printf("Error getting suppliers: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var suppliers []map[string]interface{}
	for cursor.Next(context.Background()) {
		var supplier map[string]interface{}
		if err := cursor.Decode(&supplier); err != nil {
			fmt.Printf("Error decoding supplier: %v\n", err)
			continue
		}
		suppliers = append(suppliers, supplier)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suppliers)
}

func (h *Handler) GetAllTokens(w http.ResponseWriter, r *http.Request) {
	// Query all tokens from database
	cursor, err := h.db.Collection("tokens").Find(context.Background(), bson.M{})
	if err != nil {
		fmt.Printf("Error getting tokens: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var tokens []map[string]interface{}
	for cursor.Next(context.Background()) {
		var token map[string]interface{}
		if err := cursor.Decode(&token); err != nil {
			fmt.Printf("Error decoding token: %v\n", err)
			continue
		}
		tokens = append(tokens, token)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

func (h *Handler) GetBalancesByAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountId := vars["accountId"]

	// Query balances for account from database
	cursor, err := h.db.Collection("balances").Find(context.Background(), bson.M{"userId": accountId})
	if err != nil {
		fmt.Printf("Error getting balances for account: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var balances []map[string]interface{}
	for cursor.Next(context.Background()) {
		var balance map[string]interface{}
		if err := cursor.Decode(&balance); err != nil {
			fmt.Printf("Error decoding balance: %v\n", err)
			continue
		}
		balances = append(balances, balance)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balances)
}

func (h *Handler) GetBalancesByToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenId := vars["tokenId"]

	// Query balances for token from database
	cursor, err := h.db.Collection("balances").Find(context.Background(), bson.M{"tokenId": tokenId})
	if err != nil {
		fmt.Printf("Error getting balances for token: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var balances []map[string]interface{}
	for cursor.Next(context.Background()) {
		var balance map[string]interface{}
		if err := cursor.Decode(&balance); err != nil {
			fmt.Printf("Error decoding balance: %v\n", err)
			continue
		}
		balances = append(balances, balance)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balances)
}

func (h *Handler) UpdateBlockHashes(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// Helper methods
func (h *Handler) logEvent(eventType, contractId, actorId string, extra map[string]interface{}) {
	eventId := generateEventID()
	event := map[string]interface{}{
		"_id":        eventId,
		"eventType":  eventType,
		"contractId": contractId,
		"actorId":    actorId,
		"timestamp":  time.Now(),
		"peerId":     h.peerID,
	}

	if extra != nil {
		for k, v := range extra {
			event[k] = v
		}
	}

	_, err := h.db.Collection("events").InsertOne(context.Background(), event)
	if err != nil {
		fmt.Printf("Failed to log event: %v\n", err)
	}

	// Create block
	blockNumber := h.getNextBlockNumber()
	blockHash := h.calculateBlockHash(blockNumber, time.Now(), h.getPreviousBlockHash(), []string{eventId})

	block := map[string]interface{}{
		"_id":          generateEventID(),
		"blockNumber":  blockNumber,
		"hash":         blockHash,
		"previousHash": h.getPreviousBlockHash(),
		"events":       []string{eventId},
		"timestamp":    time.Now(),
		"peerId":       h.peerID,
	}

	_, err = h.db.Collection("blocks").InsertOne(context.Background(), block)
	if err != nil {
		fmt.Printf("Failed to create block: %v\n", err)
	}
}

func (h *Handler) getPreviousBlockHash() string {
	var lastBlock map[string]interface{}
	err := h.db.Collection("blocks").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"blockNumber": -1})).Decode(&lastBlock)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "genesis"
		}
		fmt.Printf("Error getting previous block hash: %v\n", err)
		return ""
	}
	if hash, ok := lastBlock["hash"].(string); ok {
		return hash
	}
	return ""
}

func (h *Handler) getNextBlockNumber() int64 {
	var lastBlock map[string]interface{}
	err := h.db.Collection("blocks").FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"blockNumber": -1})).Decode(&lastBlock)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 1
		}
		fmt.Printf("Error getting next block number: %v\n", err)
		return 1
	}
	if blockNum, ok := lastBlock["blockNumber"].(int64); ok {
		return blockNum + 1
	}
	return 1
}

func (h *Handler) calculateBlockHash(blockNumber int64, timestamp time.Time, previousHash string, events []string) string {
	data := fmt.Sprintf("%d%s%s%v", blockNumber, timestamp.String(), previousHash, events)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
