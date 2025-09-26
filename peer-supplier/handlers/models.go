package handlers

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Balance represents a user balance
type Balance struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TokenId          string             `bson:"tokenId" json:"tokenId"`
	Account          string             `bson:"account" json:"account"`
	Balance          float64            `bson:"balance" json:"balance"`
	TransferredFrom  string             `bson:"transferredFrom,omitempty" json:"transferredFrom,omitempty"`
	CreatedAt        time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt        time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Token represents a blockchain token
type Token struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ContractId string             `bson:"contractId" json:"contractId"`
	Symbol     string             `bson:"symbol" json:"symbol"`
	Total      float64            `bson:"total" json:"total"`
	Issuer     string             `bson:"issuer" json:"issuer"`
	Owner      string             `bson:"owner" json:"owner"`
	CreatedAt  string             `bson:"createdAt" json:"createdAt"`
}
