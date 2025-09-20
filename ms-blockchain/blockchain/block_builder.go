package blockchain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"ms-blockchain/models"
)

type BlockBuilder struct {
	db            *mongo.Database
	maxTxPerBlock int
	blockInterval time.Duration
}

func NewBlockBuilder(db *mongo.Database) *BlockBuilder {
	return &BlockBuilder{
		db:            db,
		maxTxPerBlock: 10,
		blockInterval: 30 * time.Second,
	}
}

func (b *BlockBuilder) Start() {
	go b.buildBlocksPeriodically()
}

func (b *BlockBuilder) buildBlocksPeriodically() {
	ticker := time.NewTicker(b.blockInterval)
	defer ticker.Stop()

	for range ticker.C {
		b.buildNextBlock()
	}
}

func (b *BlockBuilder) buildNextBlock() {
	// Get unincluded transactions
	cur, err := b.db.Collection("transactions").Find(
		context.Background(),
		bson.M{"included": false},
		options.Find().SetLimit(int64(b.maxTxPerBlock)),
	)
	if err != nil {
		fmt.Printf("Error getting transactions: %v\n", err)
		return
	}
	defer cur.Close(context.Background())

	var transactions []models.Transaction
	if err := cur.All(context.Background(), &transactions); err != nil {
		fmt.Printf("Error decoding transactions: %v\n", err)
		return
	}

	if len(transactions) == 0 {
		return // No transactions to process
	}

	// Get latest block
	var latestBlock models.Block
	err = b.db.Collection("blocks").FindOne(
		context.Background(),
		bson.M{},
		options.FindOne().SetSort(bson.M{"block_number": -1}),
	).Decode(&latestBlock)

	var blockNumber int64 = 1
	var previousHash string = ""
	if err != mongo.ErrNoDocuments {
		blockNumber = latestBlock.BlockNumber + 1
		previousHash = latestBlock.Hash
	}

	// Create block
	txIDs := make([]string, len(transactions))
	var dataToHash string
	for i, tx := range transactions {
		txIDs[i] = tx.ID.Hex()
		dataToHash += tx.ID.Hex()
	}
	dataToHash += previousHash

	hash := sha256.Sum256([]byte(dataToHash))
	block := models.Block{
		BlockNumber:  blockNumber,
		Timestamp:    time.Now(),
		PreviousHash: previousHash,
		Hash:         hex.EncodeToString(hash[:]),
		TxIDs:        txIDs,
	}

	// Save block
	if _, err := b.db.Collection("blocks").InsertOne(context.Background(), block); err != nil {
		fmt.Printf("Error saving block: %v\n", err)
		return
	}

	// Mark transactions as included
	_, err = b.db.Collection("transactions").UpdateMany(
		context.Background(),
		bson.M{"_id": bson.M{"$in": transactions}},
		bson.M{"$set": bson.M{"included": true}},
	)
	if err != nil {
		fmt.Printf("Error updating transactions: %v\n", err)
	}
}
