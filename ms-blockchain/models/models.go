package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	StatusPending             = "PENDING"
	StatusPartiallyApproved   = "PARTIALLY_APPROVED"
	StatusApproved            = "APPROVED"
	StatusApprovedPendingExec = "APPROVED_PENDING_EXEC"
	StatusExecuted            = "EXECUTED"
	StatusFailed              = "FAILED"
)

type Transaction struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TransactionID string             `bson:"transaction_id" json:"transaction_id"`
	FromAccount   string             `bson:"from_account" json:"from_account"`
	ToAccount     string             `bson:"to_account" json:"to_account"`
	Amount        float64            `bson:"amount" json:"amount"`
	Status        string             `bson:"status" json:"status"`
	Type          string             `bson:"type" json:"type"` // CREATE, APPROVE, EXECUTE
	ApproverID    string             `bson:"approver_id,omitempty" json:"approver_id,omitempty"`
	Timestamp     time.Time          `bson:"timestamp" json:"timestamp"`
	Included      bool               `bson:"included" json:"included"`
}

type Block struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BlockNumber  int64              `bson:"block_number" json:"block_number"`
	Timestamp    time.Time          `bson:"timestamp" json:"timestamp"`
	PreviousHash string             `bson:"previous_hash" json:"previous_hash"`
	Hash         string             `bson:"hash" json:"hash"`
	Transactions []Transaction      `bson:"transactions" json:"transactions"`
}

type Approver struct {
	UserID    string    `bson:"user_id" json:"user_id"`
	Status    string    `bson:"status" json:"status"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

type WorldState struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TransactionID string             `bson:"transaction_id" json:"transaction_id"`
	FromAccount   string             `bson:"from_account" json:"from_account"`
	ToAccount     string             `bson:"to_account" json:"to_account"`
	Amount        float64            `bson:"amount" json:"amount"`
	Status        string             `bson:"status" json:"status"`
	Approvers     []Approver         `bson:"approvers" json:"approvers"`
	ApprovalCount int                `bson:"approval_count" json:"approval_count"`
	SupplierRef   string             `bson:"supplier_ref,omitempty" json:"supplier_ref,omitempty"`
	LastUpdated   time.Time          `bson:"last_updated" json:"last_updated"`
}
