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

	"peer-main-bank/chaincodeclient"
)

type Handler struct {
	db              *mongo.Database
	chaincodeClient *chaincodeclient.ChaincodeClient
	peerID          string
}

func generateEventID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func NewHandler(db *mongo.Database) *Handler {
	peerID := os.Getenv("PEER_NODE_ID")
	if peerID == "" {
		peerID = "peer-anchor"
	}

	chaincodeClient := chaincodeclient.NewClient()

	return &Handler{
		db:              db,
		chaincodeClient: chaincodeClient,
		peerID:          peerID,
	}
}

// CreateContract creates a new contract (Anchor) - Using Chaincode Service
func (h *Handler) CreateContract(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Description string   `json:"description"`
		AnchorId    string   `json:"anchorId"`
		BankId      string   `json:"bankId"`
		Amount      float64  `json:"amount"`
		Suppliers   []string `json:"suppliers"`
		FileHash    string   `json:"fileHash"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call chaincode service
	resp, err := h.chaincodeClient.InvokeCreateContract(
		req.AnchorId,
		req.Suppliers,
		req.Amount,
		req.FileHash,
	)
	if err != nil {
		fmt.Printf("Failed to create contract: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	// Log event locally
	h.logEvent("CONTRACT_CREATED", resp.ContractId, req.AnchorId, nil)
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

	resp, err := h.chaincodeClient.InvokeApproveContract(contractId, req.SupplierId)
	if err != nil {
		fmt.Printf("Failed to approve contract: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	h.logEvent("CONTRACT_APPROVED", contractId, req.SupplierId, nil)
}

// ApproveContractByBank handles bank approval of contracts
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

	// Call chaincode to issue token
	resp, err := h.chaincodeClient.InvokeIssueToken(contractId, req.BankId, 50000.0)
	if err != nil {
		fmt.Printf("Failed to issue token: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	h.logEvent("BANK_APPROVED_TOKEN_GENERATED", contractId, req.BankId, map[string]interface{}{
		"tokenId": resp.TokenId,
		"amount":  50000.0,
	})
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
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *Handler) TransferToken(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *Handler) SettleToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TokenId    string `json:"tokenId"`
		SupplierId string `json:"supplierId"`
		BankId     string `json:"bankId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.chaincodeClient.InvokeSettleToken(req.TokenId, req.SupplierId, req.BankId)
	if err != nil {
		fmt.Printf("Failed to settle token: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	h.logEvent("TOKEN_SETTLED", req.TokenId, req.SupplierId, map[string]interface{}{
		"bankId": req.BankId,
	})
}

func (h *Handler) GetTokensIssuedByBank(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *Handler) GetSuppliers(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *Handler) GetAllTokens(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *Handler) GetBalancesByAccount(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *Handler) GetBalancesByToken(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
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
