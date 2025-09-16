package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	StatusPending  = "PENDING"
	StatusApproved = "APPROVED"
	StatusExecuted = "EXECUTED"
	StatusFailed   = "FAILED"
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
}

type Block struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BlockNumber  int64              `bson:"block_number" json:"block_number"`
	Timestamp    time.Time          `bson:"timestamp" json:"timestamp"`
	PreviousHash string             `bson:"previous_hash" json:"previous_hash"`
	Hash         string             `bson:"hash" json:"hash"`
	Transactions []Transaction      `bson:"transactions" json:"transactions"`
}

type WorldState struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TransactionID string             `bson:"transaction_id" json:"transaction_id"`
	FromAccount   string             `bson:"from_account" json:"from_account"`
	ToAccount     string             `bson:"to_account" json:"to_account"`
	Amount        float64            `bson:"amount" json:"amount"`
	Status        string             `bson:"status" json:"status"`
	ApprovalCount int                `bson:"approval_count" json:"approval_count"`
	LastUpdated   time.Time          `bson:"last_updated" json:"last_updated"`
}
