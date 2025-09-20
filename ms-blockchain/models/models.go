package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	StatusPending             = "PENDING"
	StatusPartiallyApproved   = "PARTIALLY_APPROVED"
	StatusReadyToExecute      = "READY_TO_EXECUTE"
	StatusExecuted            = "EXECUTED"
	StatusApprovedPendingExec = "APPROVED_PENDING_EXEC"
	StatusFailed              = "FAILED"

	TxTypeCreate          = "CREATE"
	TxTypeApproveBank     = "APPROVE_BANK"
	TxTypeApproveSupplier = "APPROVE_SUPPLIER"
	TxTypeExecute         = "EXECUTE"
)

type Supplier struct {
	ID              string  `bson:"id" json:"id"`
	Name            string  `bson:"name" json:"name"`
	AllocatedAmount float64 `bson:"allocated_amount" json:"allocated_amount"`
	Status          string  `bson:"status" json:"status"`
	SupplierRef     string  `bson:"supplier_ref,omitempty" json:"supplier_ref,omitempty"`
}

type ContractEvent struct {
	EventID    string                 `bson:"event_id" json:"eventId"`
	ContractID string                 `bson:"contract_id" json:"contractId"`
	Type       string                 `bson:"type" json:"type"`
	ActorID    string                 `bson:"actor_id" json:"actorId"`
	Payload    map[string]interface{} `bson:"payload,omitempty" json:"payload,omitempty"`
	Timestamp  time.Time              `bson:"timestamp" json:"timestamp"`
	Included   bool                   `bson:"included" json:"included"`
}

type ContractEventInBlock struct {
	ContractID string                 `bson:"contract_id" json:"contractId"`
	EventID    string                 `bson:"event_id" json:"eventId"`
	Type       string                 `bson:"type" json:"type"`
	ActorID    string                 `bson:"actor_id" json:"actorId"`
	Payload    map[string]interface{} `bson:"payload,omitempty" json:"payload,omitempty"`
	Timestamp  time.Time              `bson:"timestamp" json:"timestamp"`
}

type Contract struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ContractID  string             `bson:"contract_id" json:"contractId"`
	Description string             `bson:"description" json:"description"`
	Buyer       string             `bson:"buyer" json:"buyer"`
	Suppliers   []Supplier         `bson:"suppliers" json:"suppliers"`
	TotalAmount float64            `bson:"total_amount" json:"totalAmount"`
	Status      string             `bson:"status" json:"status"`
	FileURL     string             `bson:"file_url,omitempty" json:"fileUrl,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`
	WordState   string             `bson:"word_state,omitempty" json:"wordState,omitempty"`
	History     []ContractEvent    `bson:"history,omitempty" json:"history,omitempty"`
}

type Transaction struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ContractID  string             `bson:"contract_id" json:"contract_id"`
	Type        string             `bson:"type" json:"type"`
	Buyer       string             `bson:"buyer" json:"buyer"`
	Bank        string             `bson:"bank" json:"bank"`
	Suppliers   []Supplier         `bson:"suppliers,omitempty" json:"suppliers,omitempty"`
	TotalAmount float64            `bson:"total_amount" json:"total_amount"`
	Description string             `bson:"description" json:"description"`
	ApproverID  string             `bson:"approver_id,omitempty" json:"approver_id,omitempty"`
	Status      string             `bson:"status" json:"status"`
	Timestamp   time.Time          `bson:"timestamp" json:"timestamp"`
	Included    bool               `bson:"included" json:"included"`
}

type Block struct {
	ID             primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	BlockNumber    int64                  `bson:"block_number" json:"blockNumber"`
	Timestamp      time.Time              `bson:"timestamp" json:"timestamp"`
	ContractEvents []ContractEventInBlock `bson:"contract_events" json:"contractEvents"`
	PreviousHash   string                 `bson:"previous_hash" json:"previousHash"`
	Hash           string                 `bson:"hash" json:"hash"`
	MerkleRoot     string                 `bson:"merkle_root" json:"merkleRoot"`
}

type Approver struct {
	ID        string    `bson:"id" json:"id"`
	Type      string    `bson:"type" json:"type"` // "BANK" or "SUPPLIER"
	Status    string    `bson:"status" json:"status"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

type WorldState struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ContractID  string             `bson:"contract_id" json:"contract_id"`
	Buyer       string             `bson:"buyer" json:"buyer"`
	Bank        Approver           `bson:"bank" json:"bank"`
	Suppliers   []Supplier         `bson:"suppliers" json:"suppliers"`
	TotalAmount float64            `bson:"total_amount" json:"total_amount"`
	Description string             `bson:"description" json:"description"`
	Status      string             `bson:"status" json:"status"`
	LastUpdated time.Time          `bson:"last_updated" json:"last_updated"`
}

type User struct {
	ID       string `bson:"id" json:"id"`
	Username string `bson:"username" json:"username"`
	Password string `bson:"password" json:"password"`
	Role     string `bson:"role" json:"role"` // "BUYER", "BANK", "SUPPLIER"
}

type ExecutionResult struct {
	Status      string `json:"status"`
	SupplierRef string `json:"supplier_ref"`
}
