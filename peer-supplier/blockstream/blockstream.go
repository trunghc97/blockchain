package blockstream

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type BlockStreamService struct {
	database *mongo.Database
}

func NewBlockStreamService(database *mongo.Database) *BlockStreamService {
	return &BlockStreamService{
		database: database,
	}
}

func (bss *BlockStreamService) StartBlockStreaming(ctx context.Context) {
	log.Println("Starting block streaming from Orderer...")

	// Simulate receiving blocks from Orderer
	// In real implementation, this would be a gRPC stream from Orderer
	go func() {
		blockNumber := 1
		for {
			select {
			case <-ctx.Done():
				log.Println("Block streaming stopped")
				return
			default:
				// Simulate receiving a block from Orderer
				time.Sleep(10 * time.Second)

				block := map[string]interface{}{
					"blockNumber": blockNumber,
					"timestamp":   time.Now().Unix(),
					"hash":        generateBlockHash(blockNumber),
					"events":      []string{"transaction_event"},
				}

				// Commit block to private ledger
				err := bss.commitBlock(block)
				if err != nil {
					log.Printf("Failed to commit block %d: %v", blockNumber, err)
				} else {
					log.Printf("Successfully committed block %d to private ledger", blockNumber)
				}

				blockNumber++
			}
		}
	}()
}

func (bss *BlockStreamService) commitBlock(block map[string]interface{}) error {
	collection := bss.database.Collection("blocks")

	// Insert block into private ledger
	_, err := collection.InsertOne(context.TODO(), block)
	return err
}

func generateBlockHash(blockNumber int) string {
	// Simple hash generation for simulation
	return fmt.Sprintf("block_hash_%d_%d", blockNumber, time.Now().Unix())
}
