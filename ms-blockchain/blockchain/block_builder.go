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
	// Get unincluded contract events
	pipeline := mongo.Pipeline{
		{{"$unwind", "$history"}},
		{{"$match", bson.M{"history.included": false}}},
		{{"$project", bson.M{
			"_id":         0,
			"contract_id": "$contract_id",
			"event_id":    "$history.event_id",
			"type":        "$history.type",
			"actor_id":    "$history.actor_id",
			"payload":     "$history.payload",
			"timestamp":   "$history.timestamp",
		}}},
		{{"$limit", b.maxTxPerBlock}},
	}

	cur, err := b.db.Collection("contracts").Aggregate(context.Background(), pipeline)
	if err != nil {
		fmt.Printf("Error getting contract events: %v\n", err)
		return
	}
	defer cur.Close(context.Background())

	var events []bson.M
	if err := cur.All(context.Background(), &events); err != nil {
		fmt.Printf("Error decoding events: %v\n", err)
		return
	}

	if len(events) == 0 {
		fmt.Printf("No pending events found\n")
		return
	}

	fmt.Printf("Found %d pending events to process\n", len(events))

	// Get latest block
	var latestBlock models.Block
	err = b.db.Collection("blocks").FindOne(
		context.Background(),
		bson.M{},
		options.FindOne().SetSort(bson.M{"block_number": -1}),
	).Decode(&latestBlock)

	var blockNumber int64 = 1
	var previousHash string = "0" // Genesis block
	if err != mongo.ErrNoDocuments {
		blockNumber = latestBlock.BlockNumber + 1
		previousHash = latestBlock.Hash
	}

	// Convert events to ContractEventInBlock
	var contractEvents []models.ContractEventInBlock
	var eventIds []string
	timestamp := time.Now()

	for _, event := range events {
		contractEvent := models.ContractEventInBlock{
			ContractID: event["contract_id"].(string),
			EventID:    event["event_id"].(string),
			Type:       event["type"].(string),
			ActorID:    event["actor_id"].(string),
			Timestamp:  timestamp,
		}

		// Handle payload if exists
		if payload, ok := event["payload"]; ok && payload != nil {
			if payloadMap, ok := payload.(bson.M); ok {
				payloadInterface := make(map[string]interface{})
				for k, v := range payloadMap {
					payloadInterface[k] = v
				}
				contractEvent.Payload = payloadInterface
			}
		}

		contractEvents = append(contractEvents, contractEvent)
		eventIds = append(eventIds, event["event_id"].(string))
	}

	// Calculate merkle root
	merkleRoot := b.calculateMerkleRoot(eventIds)

	// Calculate block hash: SHA-256(prevHash + merkleRoot + timestamp)
	hashInput := previousHash + merkleRoot + fmt.Sprintf("%d", timestamp.Unix())
	blockHash := sha256.Sum256([]byte(hashInput))

	block := models.Block{
		BlockNumber:    blockNumber,
		Timestamp:      timestamp,
		ContractEvents: contractEvents,
		PreviousHash:   previousHash,
		Hash:           hex.EncodeToString(blockHash[:]),
		MerkleRoot:     merkleRoot,
	}

	// Save block
	_, err = b.db.Collection("blocks").InsertOne(context.Background(), block)
	if err != nil {
		fmt.Printf("Error saving block: %v\n", err)
		return
	}

	fmt.Printf("Created block #%d with hash: %s\n", blockNumber, block.Hash)

	// Mark events as included
	for _, eventId := range eventIds {
		filter := bson.M{
			"history.event_id": eventId,
		}
		update := bson.M{
			"$set": bson.M{
				"history.$.included": true,
			},
		}

		_, err = b.db.Collection("contracts").UpdateOne(context.Background(), filter, update)
		if err != nil {
			fmt.Printf("Error updating event %s: %v\n", eventId, err)
		}
	}

	fmt.Printf("Marked %d events as included in block #%d\n", len(eventIds), blockNumber)
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
