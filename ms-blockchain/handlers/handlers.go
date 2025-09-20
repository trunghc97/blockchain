package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"ms-blockchain/models"
)

type Handler struct {
	db *mongo.Database
}

func NewHandler(db *mongo.Database) *Handler {
	return &Handler{db: db}
}

func (h *Handler) CreateContract(w http.ResponseWriter, r *http.Request) {
	var tx models.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx.Type = models.TxTypeCreate
	tx.Status = models.StatusPending
	tx.Timestamp = time.Now()

	// Create transaction
	txResult, err := h.db.Collection("transactions").InsertOne(context.Background(), tx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Initialize world state
	approvers := make([]models.Approver, len(tx.Suppliers)+1)
	approvers[0] = models.Approver{
		ID:        tx.Bank,
		Type:      "BANK",
		Status:    models.StatusPending,
		Timestamp: time.Now(),
	}

	for i, supplier := range tx.Suppliers {
		approvers[i+1] = models.Approver{
			ID:        supplier.ID,
			Type:      "SUPPLIER",
			Status:    models.StatusPending,
			Timestamp: time.Now(),
		}
		supplier.Status = models.StatusPending
	}

	worldState := models.WorldState{
		ContractID:  tx.ContractID,
		Buyer:       tx.Buyer,
		Bank:        approvers[0],
		Suppliers:   tx.Suppliers,
		TotalAmount: tx.TotalAmount,
		Description: tx.Description,
		Status:      models.StatusPending,
		LastUpdated: time.Now(),
	}

	if _, err := h.db.Collection("world_state").InsertOne(context.Background(), worldState); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"transaction_id": txResult.InsertedID,
		"status":         "success",
	})
}

func (h *Handler) ApproveContract(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ContractID string `json:"contract_id"`
		ApproverID string `json:"approver_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get current world state
	var worldState models.WorldState
	err := h.db.Collection("world_state").FindOne(context.Background(), bson.M{
		"contract_id": req.ContractID,
	}).Decode(&worldState)
	if err != nil {
		http.Error(w, "Contract not found", http.StatusNotFound)
		return
	}

	// Create approval transaction
	tx := models.Transaction{
		ContractID: req.ContractID,
		ApproverID: req.ApproverID,
		Timestamp:  time.Now(),
	}

	// Update approver status
	if req.ApproverID == worldState.Bank.ID {
		tx.Type = models.TxTypeApproveBank
		worldState.Bank.Status = models.StatusReadyToExecute
		worldState.Bank.Timestamp = time.Now()
	} else {
		tx.Type = models.TxTypeApproveSupplier
		for i, supplier := range worldState.Suppliers {
			if supplier.ID == req.ApproverID {
				worldState.Suppliers[i].Status = models.StatusReadyToExecute
				break
			}
		}
	}

	// Check if all approved
	allApproved := worldState.Bank.Status == models.StatusReadyToExecute
	if allApproved {
		for _, supplier := range worldState.Suppliers {
			if supplier.Status != models.StatusReadyToExecute {
				allApproved = false
				break
			}
		}
	}

	if allApproved {
		worldState.Status = models.StatusReadyToExecute
		go h.executeContract(worldState) // Trigger execution
	} else {
		worldState.Status = models.StatusPartiallyApproved
	}

	// Save transaction
	if _, err := h.db.Collection("transactions").InsertOne(context.Background(), tx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update world state
	worldState.LastUpdated = time.Now()
	if _, err := h.db.Collection("world_state").ReplaceOne(
		context.Background(),
		bson.M{"contract_id": req.ContractID},
		worldState,
	); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) executeContract(worldState models.WorldState) {
	tx := models.Transaction{
		ContractID: worldState.ContractID,
		Type:       models.TxTypeExecute,
		Timestamp:  time.Now(),
	}

	// Mock execution for each supplier
	allSuccess := true
	for i := range worldState.Suppliers {
		// Mock external call
		result := h.mockExecuteSupplierFunding()

		if result.Status == "SUCCESS" {
			worldState.Suppliers[i].SupplierRef = result.SupplierRef
			worldState.Suppliers[i].Status = models.StatusExecuted
		} else {
			allSuccess = false
			worldState.Suppliers[i].Status = models.StatusFailed
		}
	}

	if allSuccess {
		worldState.Status = models.StatusExecuted
		tx.Status = models.StatusExecuted
	} else {
		worldState.Status = models.StatusApprovedPendingExec
		tx.Status = models.StatusFailed
	}

	// Save execution transaction
	h.db.Collection("transactions").InsertOne(context.Background(), tx)

	// Update world state
	worldState.LastUpdated = time.Now()
	h.db.Collection("world_state").ReplaceOne(
		context.Background(),
		bson.M{"contract_id": worldState.ContractID},
		worldState,
	)
}

func (h *Handler) mockExecuteSupplierFunding() models.ExecutionResult {
	// Mock success with 90% probability
	if rand.Float32() < 0.9 {
		return models.ExecutionResult{
			Status:      "SUCCESS",
			SupplierRef: fmt.Sprintf("SCF-%d", rand.Int31()),
		}
	}
	return models.ExecutionResult{
		Status: "FAILED",
	}
}

func (h *Handler) QueryLedger(w http.ResponseWriter, r *http.Request) {
	contractID := r.URL.Query().Get("contract_id")
	if contractID == "" {
		http.Error(w, "contract_id is required", http.StatusBadRequest)
		return
	}

	// Get transactions
	cur, err := h.db.Collection("transactions").Find(context.Background(), bson.M{
		"contract_id": contractID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.Background())

	var transactions []models.Transaction
	if err := cur.All(context.Background(), &transactions); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get blocks containing these transactions
	var blocks []models.Block
	if len(transactions) > 0 {
		cur, err = h.db.Collection("blocks").Find(context.Background(), bson.M{
			"tx_ids": bson.M{
				"$in": transactions,
			},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cur.Close(context.Background())

		if err := cur.All(context.Background(), &blocks); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"transactions": transactions,
		"blocks":       blocks,
	})
}

func (h *Handler) ListContracts(w http.ResponseWriter, r *http.Request) {
	cur, err := h.db.Collection("world_state").Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.Background())

	var contracts []models.WorldState
	if err := cur.All(context.Background(), &contracts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(contracts)
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	cur, err := h.db.Collection("users").Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.Background())

	var users []models.User
	if err := cur.All(context.Background(), &users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}
