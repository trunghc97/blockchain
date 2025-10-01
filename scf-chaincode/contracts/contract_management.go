package contracts

import "time"

// Contract represents a supply chain finance contract
type Contract struct {
	ID          string    `bson:"_id"`
	AnchorID    string    `bson:"anchorId"`
	Suppliers   []string  `bson:"suppliers"`
	TotalAmount float64   `bson:"totalAmount"`
	Status      string    `bson:"status"` // PENDING | APPROVED | EXECUTED
	FileHash    string    `bson:"fileHash"`
	CreatedAt   time.Time `bson:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt"`
}

// CreateContract creates a new contract
func CreateContract(anchorID string, suppliers []string, totalAmount float64, fileHash string) (*Contract, error) {
	contract := &Contract{
		ID:          "contract-" + anchorID,
		AnchorID:    anchorID,
		Suppliers:   suppliers,
		TotalAmount: totalAmount,
		Status:      "PENDING",
		FileHash:    fileHash,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// TODO: Save to database

	return contract, nil
}

// ApproveContract approves a contract by a supplier
func ApproveContract(contractID, supplierID string) (*Contract, error) {
	// TODO: Load contract from database and update supplier approval status

	contract := &Contract{
		ID:        contractID,
		Status:    "APPROVED",
		UpdatedAt: time.Now(),
	}

	return contract, nil
}

// FinalizeContract finalizes a contract when all suppliers have approved
func FinalizeContract(contractID string) (*Contract, error) {
	// TODO: Load contract and check all suppliers approved

	contract := &Contract{
		ID:        contractID,
		Status:    "EXECUTED",
		UpdatedAt: time.Now(),
	}

	return contract, nil
}
