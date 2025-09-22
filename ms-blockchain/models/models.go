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
	TxTypeApproveSupplier = "APPROVE_SUPPLIER"
	TxTypeExecute         = "EXECUTE"
)

type Supplier struct {
	SupplierID      string  `bson:"supplierId" json:"supplierId"`
	Name            string  `bson:"name" json:"name"`
	AllocatedAmount float64 `bson:"allocatedAmount" json:"allocatedAmount"`
	Status          string  `bson:"status" json:"status"`
	SupplierRef     string  `bson:"supplierRef,omitempty" json:"supplierRef,omitempty"`
}

type ContractEvent struct {
	EventID    string                 `bson:"eventId" json:"eventId"`
	ContractID string                 `bson:"contractId" json:"contractId"`
	Type       string                 `bson:"type" json:"type"`
	ActorID    string                 `bson:"actorId" json:"actorId"`
	Payload    map[string]interface{} `bson:"payload,omitempty" json:"payload,omitempty"`
	Timestamp  time.Time              `bson:"timestamp" json:"timestamp"`
	Included   bool                   `bson:"included" json:"included"`
}

type BlockEvent struct {
	EventID    string                 `bson:"eventId" json:"eventId"`
	ContractID string                 `bson:"contractId" json:"contractId"`
	Type       string                 `bson:"type" json:"type"`
	Payload    map[string]interface{} `bson:"payload,omitempty" json:"payload,omitempty"`
	Timestamp  time.Time              `bson:"timestamp" json:"timestamp"`
}

type Contract struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ContractID  string             `bson:"contractId" json:"contractId"`
	Description string             `bson:"description" json:"description"`
	Buyer       string             `bson:"buyer" json:"buyer"`
	Suppliers   []Supplier         `bson:"suppliers" json:"suppliers"`
	TotalAmount float64            `bson:"totalAmount" json:"totalAmount"`
	Status      string             `bson:"status" json:"status"`
	FileURL     string             `bson:"fileUrl,omitempty" json:"fileUrl,omitempty"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
	WordState   string             `bson:"wordState,omitempty" json:"wordState,omitempty"`
	History     []ContractEvent    `bson:"history,omitempty" json:"history,omitempty"`
}

type Block struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BlockNumber int64              `bson:"blockNumber" json:"blockNumber"`
	Timestamp   time.Time          `bson:"timestamp" json:"timestamp"`
	Events      []BlockEvent       `bson:"events" json:"events"`
	PrevHash    string             `bson:"prevHash" json:"prevHash"`
	Hash        string             `bson:"hash" json:"hash"`
	MerkleRoot  string             `bson:"merkleRoot" json:"merkleRoot"`
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
