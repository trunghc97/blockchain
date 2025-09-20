package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
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
	var req struct {
		ContractID  string            `json:"contractId"`
		Description string            `json:"description"`
		Buyer       string            `json:"buyer"`
		Suppliers   []models.Supplier `json:"suppliers"`
		TotalAmount float64           `json:"totalAmount"`
		FileURL     string            `json:"fileUrl,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set supplier statuses to PENDING
	for i := range req.Suppliers {
		req.Suppliers[i].Status = models.StatusPending
	}

	// Generate unique contract ID if not provided
	if req.ContractID == "" {
		bytes := make([]byte, 16)
		rand.Read(bytes)
		req.ContractID = hex.EncodeToString(bytes)
	}

	// Create CREATE event
	eventId := h.generateEventID()
	createEvent := models.ContractEvent{
		EventID:    eventId,
		ContractID: req.ContractID,
		Type:       "CREATE",
		ActorID:    req.Buyer,
		Timestamp:  time.Now(),
		Included:   false,
	}

	contract := models.Contract{
		ContractID:  req.ContractID,
		Description: req.Description,
		Buyer:       req.Buyer,
		Suppliers:   req.Suppliers,
		TotalAmount: req.TotalAmount,
		Status:      models.StatusPending,
		FileURL:     req.FileURL,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		History:     []models.ContractEvent{createEvent},
	}

	// Insert contract
	result, err := h.db.Collection("contracts").InsertOne(context.Background(), contract)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"contract_id": contract.ContractID,
		"id":          result.InsertedID,
		"status":      "success",
	})
}

func (h *Handler) generateEventID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (h *Handler) ApproveContract(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ContractID string `json:"contractId"`
		SupplierID string `json:"supplierId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get contract
	var contract models.Contract
	err := h.db.Collection("contracts").FindOne(context.Background(), bson.M{
		"contract_id": req.ContractID,
	}).Decode(&contract)
	if err != nil {
		http.Error(w, "Contract not found", http.StatusNotFound)
		return
	}

	// Find and update supplier status
	supplierFound := false
	for i, supplier := range contract.Suppliers {
		if supplier.ID == req.SupplierID {
			contract.Suppliers[i].Status = models.StatusReadyToExecute
			supplierFound = true
			break
		}
	}

	if !supplierFound {
		http.Error(w, "Supplier not found in contract", http.StatusBadRequest)
		return
	}

	// Check if all suppliers approved
	allApproved := true
	for _, supplier := range contract.Suppliers {
		if supplier.Status != models.StatusReadyToExecute {
			allApproved = false
			break
		}
	}

	// Create APPROVE_SUPPLIER event
	eventId := h.generateEventID()
	payload := map[string]interface{}{
		"supplierId":  req.SupplierID,
		"allApproved": allApproved,
	}

	approveEvent := models.ContractEvent{
		EventID:    eventId,
		ContractID: req.ContractID,
		Type:       "APPROVE_SUPPLIER",
		ActorID:    req.SupplierID,
		Payload:    payload,
		Timestamp:  time.Now(),
		Included:   false,
	}

	// Add event to history
	contract.History = append(contract.History, approveEvent)
	contract.UpdatedAt = time.Now()

	// Update contract status if all approved
	if allApproved {
		contract.Status = models.StatusReadyToExecute
		// Trigger execution
		go h.executeContract(req.ContractID)
	}

	// Update contract
	if _, err := h.db.Collection("contracts").ReplaceOne(
		context.Background(),
		bson.M{"contract_id": req.ContractID},
		contract,
	); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) executeContract(contractID string) {
	// Get contract
	var contract models.Contract
	err := h.db.Collection("contracts").FindOne(context.Background(), bson.M{
		"contract_id": contractID,
	}).Decode(&contract)
	if err != nil {
		fmt.Printf("Error getting contract for execution: %v\n", err)
		return
	}

	// Mock execution for each supplier
	allSuccess := true
	var executedSuppliers []string
	var failedSuppliers []string

	for i := range contract.Suppliers {
		// Mock external call
		result := h.mockExecuteSupplierFunding()

		if result.Status == "SUCCESS" {
			contract.Suppliers[i].SupplierRef = result.SupplierRef
			contract.Suppliers[i].Status = models.StatusExecuted
			executedSuppliers = append(executedSuppliers, contract.Suppliers[i].ID)
		} else {
			allSuccess = false
			contract.Suppliers[i].Status = models.StatusFailed
			failedSuppliers = append(failedSuppliers, contract.Suppliers[i].ID)
		}
	}

	// Update contract status
	if allSuccess {
		contract.Status = models.StatusExecuted
	} else {
		contract.Status = models.StatusApprovedPendingExec
	}

	// Create EXECUTE event
	eventId := h.generateEventID()
	payload := map[string]interface{}{
		"executedSuppliers": executedSuppliers,
		"failedSuppliers":   failedSuppliers,
		"allSuccess":        allSuccess,
	}

	executeEvent := models.ContractEvent{
		EventID:    eventId,
		ContractID: contractID,
		Type:       "EXECUTE",
		ActorID:    "SYSTEM", // System triggered execution
		Payload:    payload,
		Timestamp:  time.Now(),
		Included:   false,
	}

	// Add event to history
	contract.History = append(contract.History, executeEvent)
	contract.UpdatedAt = time.Now()

	// Update contract
	if _, err := h.db.Collection("contracts").ReplaceOne(
		context.Background(),
		bson.M{"contract_id": contractID},
		contract,
	); err != nil {
		fmt.Printf("Error updating contract after execution: %v\n", err)
	}
}

func (h *Handler) mockExecuteSupplierFunding() models.ExecutionResult {
	// Mock success - always success for demo
	return models.ExecutionResult{
		Status:      "SUCCESS",
		SupplierRef: fmt.Sprintf("SCF-%d", time.Now().Unix()),
	}
}

func (h *Handler) QueryLedger(w http.ResponseWriter, r *http.Request) {
	contractID := r.URL.Query().Get("contract_id")
	if contractID == "" {
		http.Error(w, "contract_id is required", http.StatusBadRequest)
		return
	}

	// Get blocks containing events for this contract
	cur, err := h.db.Collection("blocks").Find(context.Background(), bson.M{
		"contract_events.contract_id": contractID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.Background())

	var blocks []models.Block
	if err := cur.All(context.Background(), &blocks); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get contract for additional info
	var contract models.Contract
	err = h.db.Collection("contracts").FindOne(context.Background(), bson.M{
		"contract_id": contractID,
	}).Decode(&contract)

	contractInfo := map[string]interface{}{
		"contractId":  contractID,
		"description": "",
		"status":      "UNKNOWN",
		"buyer":       "",
		"totalAmount": 0.0,
	}

	if err == nil {
		contractInfo["description"] = contract.Description
		contractInfo["status"] = contract.Status
		contractInfo["buyer"] = contract.Buyer
		contractInfo["totalAmount"] = contract.TotalAmount
	}

	// Convert blocks to transaction format for frontend compatibility
	var transactions []map[string]interface{}
	for _, block := range blocks {
		if block.ContractEvents != nil {
			for _, event := range block.ContractEvents {
				if event.ContractID == contractID {
					transaction := map[string]interface{}{
						"id":          event.EventID,
						"contractId":  event.ContractID,
						"type":        event.Type,
						"buyer":       contractInfo["buyer"],
						"bank":        "",
						"suppliers":   []interface{}{},
						"totalAmount": contractInfo["totalAmount"],
						"description": contractInfo["description"],
						"approverID":  event.ActorID,
						"status":      contractInfo["status"],
						"timestamp":   event.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
						"included":    true,
						"blockNumber": block.BlockNumber,
						"blockHash":   block.Hash,
						"merkleRoot":  block.MerkleRoot,
					}
					transactions = append(transactions, transaction)
				}
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"transactions": transactions,
		"blocks":       blocks,
		"contractId":   contractID,
	})
}

func (h *Handler) QueryContractLedger(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contractID := vars["id"]

	if contractID == "" {
		http.Error(w, "Contract ID is required", http.StatusBadRequest)
		return
	}

	// Use the same logic as QueryLedger but with contract ID from path
	h.QueryLedger(w, &http.Request{
		Method: "GET",
		URL:    &url.URL{RawQuery: "contract_id=" + contractID},
	})
}

func (h *Handler) ListContracts(w http.ResponseWriter, r *http.Request) {
	cur, err := h.db.Collection("contracts").Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.Background())

	var contracts []models.Contract
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
