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
		blockInterval: 10 * time.Second,
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
	// Get unincluded contract events from events collection
	filter := bson.M{"included": false}
	opts := options.Find().SetLimit(int64(b.maxTxPerBlock))

	cur, err := b.db.Collection("events").Find(context.Background(), filter, opts)
	if err != nil {
		fmt.Printf("Error getting contract events: %v\n", err)
		return
	}
	defer cur.Close(context.Background())

	var events []models.ContractEvent
	if err := cur.All(context.Background(), &events); err != nil {
		fmt.Printf("Error decoding events: %v\n", err)
		return
	}

	if len(events) == 0 {
		return
	}

	// Get latest block
	var latestBlock models.Block
	err = b.db.Collection("blocks").FindOne(
		context.Background(),
		bson.M{},
		options.FindOne().SetSort(bson.M{"blockNumber": -1}),
	).Decode(&latestBlock)

	var blockNumber int64 = 1
	var prevHash string = "0" // Genesis block
	if err != mongo.ErrNoDocuments {
		blockNumber = latestBlock.BlockNumber + 1
		prevHash = latestBlock.Hash
	}

	// Convert events to BlockEvent
	var blockEvents []models.BlockEvent
	var eventIds []string
	timestamp := time.Now()

	for _, event := range events {
		blockEvent := models.BlockEvent{
			EventID:    event.EventID,
			ContractID: event.ContractID,
			Type:       event.Type,
			Payload:    event.Payload,
			Timestamp:  event.Timestamp, // Use original event timestamp
		}

		blockEvents = append(blockEvents, blockEvent)
		eventIds = append(eventIds, event.EventID)
	}

	// Calculate merkle root
	merkleRoot := b.calculateMerkleRoot(eventIds)

	// Calculate block hash: SHA256(prevHash + merkleRoot + timestamp)
	hashInput := prevHash + merkleRoot + fmt.Sprintf("%d", timestamp.Unix())
	blockHash := sha256.Sum256([]byte(hashInput))

	block := models.Block{
		BlockNumber: blockNumber,
		Timestamp:   timestamp,
		Events:      blockEvents,
		PrevHash:    prevHash,
		Hash:        hex.EncodeToString(blockHash[:]),
		MerkleRoot:  merkleRoot,
	}

	// Save block
	_, err = b.db.Collection("blocks").InsertOne(context.Background(), block)
	if err != nil {
		fmt.Printf("Error saving block: %v\n", err)
		return
	}

	// Mark events as included
	for _, eventId := range eventIds {
		filter := bson.M{
			"eventId": eventId,
		}
		update := bson.M{
			"$set": bson.M{
				"included": true,
			},
		}

		_, err = b.db.Collection("events").UpdateOne(context.Background(), filter, update)
		if err != nil {
			fmt.Printf("Error updating event %s: %v\n", eventId, err)
		}
	}
}

func (b *BlockBuilder) calculateMerkleRoot(eventIds []string) string {
	if len(eventIds) == 0 {
		return b.calculateSHA256("")
	}

	// Calculate SHA256 for each event ID
	hashes := make([]string, len(eventIds))
	for i, eventId := range eventIds {
		hashes[i] = b.calculateSHA256(eventId)
	}

	// Build merkle tree
	for len(hashes) > 1 {
		var newHashes []string
		for i := 0; i < len(hashes); i += 2 {
			left := hashes[i]
			right := ""
			if i+1 < len(hashes) {
				right = hashes[i+1]
			} else {
				right = left // Duplicate last hash if odd number
			}
			newHashes = append(newHashes, b.calculateSHA256(left+right))
		}
		hashes = newHashes
	}

	return hashes[0]
}

func (b *BlockBuilder) calculateSHA256(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
