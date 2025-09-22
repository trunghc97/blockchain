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

	// Insert contract event only (remove contract storage from blockchain service)
	eventDoc := bson.M{
		"eventId":    createEvent.EventID,
		"contractId": createEvent.ContractID,
		"type":       createEvent.Type,
		"actorId":    createEvent.ActorID,
		"payload":    createEvent.Payload,
		"timestamp":  createEvent.Timestamp,
		"included":   false,
	}

	_, err := h.db.Collection("events").InsertOne(context.Background(), eventDoc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"contractId": contract.ContractID,
		"status":     "success",
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
		"contractId": req.ContractID,
	}).Decode(&contract)
	if err != nil {
		http.Error(w, "Contract not found", http.StatusNotFound)
		return
	}

	// Find and update supplier status
	supplierFound := false
	for i, supplier := range contract.Suppliers {
		if supplier.SupplierID == req.SupplierID {
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

	// Insert event to events collection
	eventDoc := bson.M{
		"eventId":    approveEvent.EventID,
		"contractId": approveEvent.ContractID,
		"type":       approveEvent.Type,
		"actorId":    approveEvent.ActorID,
		"payload":    approveEvent.Payload,
		"timestamp":  approveEvent.Timestamp,
		"included":   false,
	}

	if _, err := h.db.Collection("events").InsertOne(context.Background(), eventDoc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update contract status if all approved
	if allApproved {
		contract.Status = models.StatusReadyToExecute
		// Trigger execution
		go h.executeContract(req.ContractID)
	}

	// Update contract
	if _, err := h.db.Collection("contracts").ReplaceOne(
		context.Background(),
		bson.M{"contractId": req.ContractID},
		contract,
	); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) executeContract(contractID string) {
	// Get contract
	var contract models.Contract
	err := h.db.Collection("contracts").FindOne(context.Background(), bson.M{
		"contractId": contractID,
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
			executedSuppliers = append(executedSuppliers, contract.Suppliers[i].SupplierID)
		} else {
			allSuccess = false
			contract.Suppliers[i].Status = models.StatusFailed
			failedSuppliers = append(failedSuppliers, contract.Suppliers[i].SupplierID)
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

	// Insert event to events collection
	eventDoc := bson.M{
		"eventId":    executeEvent.EventID,
		"contractId": executeEvent.ContractID,
		"type":       executeEvent.Type,
		"actorId":    executeEvent.ActorID,
		"payload":    executeEvent.Payload,
		"timestamp":  executeEvent.Timestamp,
		"included":   false,
	}

	if _, err := h.db.Collection("events").InsertOne(context.Background(), eventDoc); err != nil {
		fmt.Printf("Error inserting execute event to events collection: %v\n", err)
		return
	}

	// Update contract
	if _, err := h.db.Collection("contracts").ReplaceOne(
		context.Background(),
		bson.M{"contractId": contractID},
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
	contractID := r.URL.Query().Get("contractId")
	if contractID == "" {
		http.Error(w, "contractId is required", http.StatusBadRequest)
		return
	}

	// Get blocks containing events for this contract
	cur, err := h.db.Collection("blocks").Find(context.Background(), bson.M{
		"events.contractId": contractID,
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
		"contractId": contractID,
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
		if block.Events != nil {
			for _, event := range block.Events {
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
						"approverID":  "", // ActorID removed from BlockEvent
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

	w.Header().Set("Content-Type", "application/json")
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

	// Get all blocks containing events for this contract
	cur, err := h.db.Collection("blocks").Find(context.Background(), bson.M{
		"events.contractId": contractID,
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

	// Extract events for this contract from all blocks
	var events []map[string]interface{}
	for _, block := range blocks {
		for _, event := range block.Events {
			if event.ContractID == contractID {
				eventInfo := map[string]interface{}{
					"eventId":     event.EventID,
					"contractId":  event.ContractID,
					"type":        event.Type,
					"payload":     event.Payload,
					"timestamp":   event.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
					"blockNumber": block.BlockNumber,
					"blockHash":   block.Hash,
					"merkleRoot":  block.MerkleRoot,
				}
				events = append(events, eventInfo)
			}
		}
	}

	// Sort events by timestamp
	// Note: In a real implementation, you'd want to sort by block number and then by event order within block

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"contractId": contractID,
		"events":     events,
		"total":      len(events),
	})
}

func (h *Handler) ListContracts(w http.ResponseWriter, r *http.Request) {
	// Note: This method may not be used anymore since contracts are managed by backend
	// Return empty array for compatibility
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]models.Contract{})
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *Handler) GetLedgerBlocks(w http.ResponseWriter, r *http.Request) {
	// Get all blocks sorted by block number
	opts := options.Find().SetSort(bson.M{"blockNumber": 1})
	cur, err := h.db.Collection("blocks").Find(context.Background(), bson.M{}, opts)
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

	// Convert to response format
	var response []map[string]interface{}
	for _, block := range blocks {
		blockInfo := map[string]interface{}{
			"blockNumber": block.BlockNumber,
			"hash":        block.Hash,
			"prevHash":    block.PrevHash,
			"merkleRoot":  block.MerkleRoot,
			"eventCount":  len(block.Events),
		}
		response = append(response, blockInfo)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
