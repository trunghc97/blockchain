package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Balance represents a user balance
type Balance struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    string             `bson:"userId" json:"userId"`
	Amount    float64            `bson:"amount" json:"amount"`
	Currency  string             `bson:"currency" json:"currency"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Token represents a blockchain token
type Token struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	From        string             `bson:"from" json:"from"`
	To          string             `bson:"to" json:"to"`
	Amount      float64            `bson:"amount" json:"amount"`
	Currency    string             `bson:"currency" json:"currency"`
	Timestamp   time.Time          `bson:"timestamp" json:"timestamp"`
	BlockHash   string             `bson:"blockHash" json:"blockHash"`
	BlockNumber int64              `bson:"blockNumber" json:"blockNumber"`
	Status      string             `bson:"status" json:"status"`
}
