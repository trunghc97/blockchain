package contracts

import "time"

// Token represents a tokenized asset
type Token struct {
	ID          string             `bson:"_id"`
	ContractID  string             `bson:"contractId"`
	Symbol      string             `bson:"symbol"`
	TotalSupply float64            `bson:"totalSupply"`
	Issuer      string             `bson:"issuer"`
	Owner       string             `bson:"owner"`
	Balances    map[string]float64 `bson:"balances"`
	CreatedAt   time.Time          `bson:"createdAt"`
}

// IssueToken issues tokens for a contract
func IssueToken(contractID, issuer string, totalSupply float64) (*Token, error) {
	token := &Token{
		ID:          "token-" + contractID,
		ContractID:  contractID,
		Symbol:      "SCF",
		TotalSupply: totalSupply,
		Issuer:      issuer,
		Owner:       issuer,
		Balances:    map[string]float64{issuer: totalSupply},
		CreatedAt:   time.Now(),
	}

	// TODO: Save to database

	return token, nil
}

// TransferToken transfers tokens between accounts
func TransferToken(tokenID, from, to string, amount float64) (*Token, error) {
	// TODO: Load token from database, validate balance, update balances

	token := &Token{
		ID: tokenID,
		Balances: map[string]float64{
			from: 0,      // Reduced amount
			to:   amount, // Increased amount
		},
	}

	return token, nil
}

// SettleToken settles tokens with the bank
func SettleToken(tokenID, supplierID, bankID string) (*Token, error) {
	// TODO: Load token, validate settlement conditions, remove supplier balance

	token := &Token{
		ID: tokenID,
		Balances: map[string]float64{
			supplierID: 0, // Settled
		},
	}

	return token, nil
}
