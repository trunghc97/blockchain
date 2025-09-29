package handlers

import (
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockPeer implements PeerInterface for main bank
type MockPeer struct{}

// PublishTransactionToKafka implements PeerInterface
func (p *MockPeer) PublishTransactionToKafka(txType, contractID, tokenID, senderID, receiverID string, amount float64) error {
	log.Printf("MockPeer: Publishing transaction to Kafka - Type: %s, Contract: %s, Token: %s, From: %s, To: %s, Amount: %f",
		txType, contractID, tokenID, senderID, receiverID, amount)
	// Mock implementation - do nothing
	return nil
}

// Balance represents a user balance
type Balance struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TokenId         string             `bson:"tokenId" json:"tokenId"`
	Account         string             `bson:"account" json:"account"`
	Balance         float64            `bson:"balance" json:"balance"`
	TransferredFrom string             `bson:"transferredFrom,omitempty" json:"transferredFrom,omitempty"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Token represents a blockchain token
type Token struct {
	ID         string  `bson:"_id" json:"id"`
	ContractId string  `bson:"contractId" json:"contractId"`
	Symbol     string  `bson:"symbol" json:"symbol"`
	Total      float64 `bson:"total" json:"total"`
	Issuer     string  `bson:"issuer" json:"issuer"`
	Owner      string  `bson:"owner" json:"owner"`
	CreatedAt  string  `bson:"createdAt" json:"createdAt"`
}
