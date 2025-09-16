package blockchain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"ms-blockchain/db"
	"ms-blockchain/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func StartBlockBuilder() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			buildBlock()
		}
	}()
}

func buildBlock() {
	txColl := db.GetCollection("transactions")
	blockColl := db.GetCollection("blocks")

	// Lấy block cuối cùng
	var lastBlock models.Block
	opts := options.FindOne().SetSort(bson.M{"block_number": -1})
	err := blockColl.FindOne(context.Background(), bson.M{}, opts).Decode(&lastBlock)
	if err != nil {
		lastBlock = models.Block{
			BlockNumber:  0,
			PreviousHash: "",
		}
	}

	// Lấy các transaction chưa được đưa vào block
	filter := bson.M{"_id": bson.M{"$nin": getBlockedTransactionIDs()}}
	cursor, err := txColl.Find(context.Background(), filter)
	if err != nil {
		log.Printf("Error getting transactions: %v", err)
		return
	}
	defer cursor.Close(context.Background())

	var transactions []models.Transaction
	if err = cursor.All(context.Background(), &transactions); err != nil {
		log.Printf("Error decoding transactions: %v", err)
		return
	}

	if len(transactions) == 0 {
		return
	}

	// Tạo block mới
	newBlock := models.Block{
		ID:           primitive.NewObjectID(),
		BlockNumber:  lastBlock.BlockNumber + 1,
		Timestamp:    time.Now(),
		PreviousHash: lastBlock.Hash,
		Transactions: transactions,
	}

	// Tính hash của block
	blockData, _ := json.Marshal(struct {
		BlockNumber  int64
		Timestamp    time.Time
		PreviousHash string
		Transactions []models.Transaction
	}{
		BlockNumber:  newBlock.BlockNumber,
		Timestamp:    newBlock.Timestamp,
		PreviousHash: newBlock.PreviousHash,
		Transactions: newBlock.Transactions,
	})

	hash := sha256.Sum256(blockData)
	newBlock.Hash = hex.EncodeToString(hash[:])

	// Lưu block mới
	_, err = blockColl.InsertOne(context.Background(), newBlock)
	if err != nil {
		log.Printf("Error saving block: %v", err)
		return
	}

	log.Printf("Created new block #%d with %d transactions", newBlock.BlockNumber, len(transactions))
}

func getBlockedTransactionIDs() []primitive.ObjectID {
	blockColl := db.GetCollection("blocks")
	cursor, err := blockColl.Find(context.Background(), bson.M{})
	if err != nil {
		return nil
	}
	defer cursor.Close(context.Background())

	var blocks []models.Block
	if err = cursor.All(context.Background(), &blocks); err != nil {
		return nil
	}

	var ids []primitive.ObjectID
	for _, block := range blocks {
		for _, tx := range block.Transactions {
			ids = append(ids, tx.ID)
		}
	}
	return ids
}
