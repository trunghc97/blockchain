package block

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"orderer-cluster/proto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// BlockBuilder creates and manages blockchain blocks
type BlockBuilder struct {
	db *mongo.Database
}

// NewBlockBuilder creates a new block builder
func NewBlockBuilder(db *mongo.Database) *BlockBuilder {
	return &BlockBuilder{db: db}
}

// CreateBlock creates a new block from transactions
func (bb *BlockBuilder) CreateBlock(transactions []*proto.Transaction, height int64) *proto.Block {
	previousHash := bb.getPreviousBlockHash()
	merkleRoot := bb.calculateMerkleRoot(transactions)
	timestamp := time.Now()

	// Create block data for hashing
	blockData := fmt.Sprintf("%d-%s-%s-%d", height, previousHash, merkleRoot, timestamp.Unix())
	hash := sha256.Sum256([]byte(blockData))
	blockHash := hex.EncodeToString(hash[:])

	protoTimestamp := timestamppb.New(timestamp)
	block := &proto.Block{
		Height:       height,
		Timestamp:    protoTimestamp, // Convert to protobuf timestamp
		Transactions: transactions,
		PreviousHash: previousHash,
		Hash:         blockHash,
		MerkleRoot:   merkleRoot,
		Signatures:   []*proto.BlockSignature{}, // Will be filled by PBFT
	}

	return block
}

// getPreviousBlockHash returns the hash of the previous block
func (bb *BlockBuilder) getPreviousBlockHash() string {
	collection := bb.db.Collection("blocks")

	var lastBlock bson.M
	err := collection.FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"height": -1})).Decode(&lastBlock)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "genesis" // Genesis block
		}
		return ""
	}

	if hash, ok := lastBlock["hash"].(string); ok {
		return hash
	}

	return ""
}

// calculateMerkleRoot calculates the Merkle root of transactions
func (bb *BlockBuilder) calculateMerkleRoot(transactions []*proto.Transaction) string {
	if len(transactions) == 0 {
		return ""
	}

	// Simple implementation - hash all transaction IDs together
	var txHashes []string
	for _, tx := range transactions {
		txHashes = append(txHashes, tx.TransactionId)
	}

	combined := ""
	for _, hash := range txHashes {
		combined += hash
	}

	rootHash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(rootHash[:])
}

// ValidateBlock validates a block's structure and transactions
func (bb *BlockBuilder) ValidateBlock(block *proto.Block) bool {
	// Verify block hash
	expectedHash := bb.calculateBlockHash(block)
	if expectedHash != block.Hash {
		return false
	}

	// Verify merkle root
	expectedMerkleRoot := bb.calculateMerkleRoot(block.Transactions)
	if expectedMerkleRoot != block.MerkleRoot {
		return false
	}

	// Additional validation can be added here
	return true
}

// calculateBlockHash calculates the hash of a block
func (bb *BlockBuilder) calculateBlockHash(block *proto.Block) string {
	blockData := fmt.Sprintf("%d-%s-%s-%d", block.Height, block.PreviousHash, block.MerkleRoot, block.Timestamp.Seconds)
	hash := sha256.Sum256([]byte(blockData))
	return hex.EncodeToString(hash[:])
}

// GetBlockByHeight retrieves a block by height
func (bb *BlockBuilder) GetBlockByHeight(height int64) (*proto.Block, error) {
	collection := bb.db.Collection("blocks")

	var block proto.Block
	err := collection.FindOne(context.Background(), bson.M{"height": height}).Decode(&block)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

// GetLatestBlockHeight returns the height of the latest block
func (bb *BlockBuilder) GetLatestBlockHeight() int64 {
	collection := bb.db.Collection("blocks")

	var lastBlock bson.M
	err := collection.FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.M{"height": -1})).Decode(&lastBlock)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0
		}
		return 0
	}

	if height, ok := lastBlock["height"].(int64); ok {
		return height
	}

	return 0
}
