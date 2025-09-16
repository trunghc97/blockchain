package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"ms-blockchain/config"
	"ms-blockchain/db"
	"ms-blockchain/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateTransaction(c *gin.Context) {
	var tx models.Transaction
	if err := c.ShouldBindJSON(&tx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx.ID = primitive.NewObjectID()
	tx.Type = "CREATE"
	tx.Status = models.StatusPending
	tx.Timestamp = time.Now()

	// Lưu transaction
	txColl := db.GetCollection("transactions")
	_, err := txColl.InsertOne(context.Background(), tx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Tạo world state
	worldState := models.WorldState{
		ID:            primitive.NewObjectID(),
		TransactionID: tx.TransactionID,
		FromAccount:   tx.FromAccount,
		ToAccount:     tx.ToAccount,
		Amount:        tx.Amount,
		Status:        models.StatusPending,
		ApprovalCount: 0,
		LastUpdated:   time.Now(),
	}

	wsColl := db.GetCollection("world_state")
	_, err = wsColl.InsertOne(context.Background(), worldState)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

func ApproveTransaction(c *gin.Context) {
	var tx models.Transaction
	if err := c.ShouldBindJSON(&tx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx.ID = primitive.NewObjectID()
	tx.Type = "APPROVE"
	tx.Status = models.StatusApproved
	tx.Timestamp = time.Now()

	// Lưu transaction approve
	txColl := db.GetCollection("transactions")
	_, err := txColl.InsertOne(context.Background(), tx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Cập nhật world state
	wsColl := db.GetCollection("world_state")
	filter := bson.M{"transaction_id": tx.TransactionID}
	update := bson.M{
		"$inc": bson.M{"approval_count": 1},
		"$set": bson.M{"last_updated": time.Now()},
	}

	var worldState models.WorldState
	err = wsColl.FindOneAndUpdate(
		context.Background(),
		filter,
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&worldState)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Kiểm tra số lượng approval
	if worldState.ApprovalCount >= config.ApprovalThreshold {
		// Gọi supplier API
		supplierReq := map[string]interface{}{
			"transaction_id": tx.TransactionID,
			"from_account":   tx.FromAccount,
			"to_account":     tx.ToAccount,
			"amount":         tx.Amount,
		}

		jsonData, _ := json.Marshal(supplierReq)
		resp, err := http.Post(config.SupplierURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			// Tạo transaction EXECUTE
			executeTx := models.Transaction{
				ID:            primitive.NewObjectID(),
				TransactionID: tx.TransactionID,
				FromAccount:   tx.FromAccount,
				ToAccount:     tx.ToAccount,
				Amount:        tx.Amount,
				Type:          "EXECUTE",
				Status:        models.StatusExecuted,
				Timestamp:     time.Now(),
			}

			_, err = txColl.InsertOne(context.Background(), executeTx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Cập nhật world state
			update = bson.M{
				"$set": bson.M{
					"status":       models.StatusExecuted,
					"last_updated": time.Now(),
				},
			}
			err = wsColl.FindOneAndUpdate(context.Background(), filter, update).Err()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	c.JSON(http.StatusOK, worldState)
}

func GetTransactionStatus(c *gin.Context) {
	txID := c.Param("id")
	if txID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction_id is required"})
		return
	}

	wsColl := db.GetCollection("world_state")
	var worldState models.WorldState
	err := wsColl.FindOne(context.Background(), bson.M{"transaction_id": txID}).Decode(&worldState)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, worldState)
}
