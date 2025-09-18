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

type CreateTransactionRequest struct {
	TransactionID string   `json:"transaction_id"`
	FromAccount   string   `json:"from_account"`
	ToAccount     string   `json:"to_account"`
	Amount        float64  `json:"amount"`
	Approvers     []string `json:"approvers"`
}

type SupplierResponse struct {
	Status      string `json:"status"`
	SupplierRef string `json:"supplier_ref"`
}

func CreateTransaction(c *gin.Context) {
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if req.TransactionID == "" || req.FromAccount == "" || req.ToAccount == "" || len(req.Approvers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
		return
	}

	// Tạo transaction CREATE
	tx := models.Transaction{
		ID:            primitive.NewObjectID(),
		TransactionID: req.TransactionID,
		FromAccount:   req.FromAccount,
		ToAccount:     req.ToAccount,
		Amount:        req.Amount,
		Type:          "CREATE",
		Status:        models.StatusPending,
		Timestamp:     time.Now(),
	}

	// Lưu transaction
	txColl := db.GetCollection("transactions")
	_, err := txColl.InsertOne(context.Background(), tx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Khởi tạo danh sách approvers
	approvers := make([]models.Approver, 0)
	for _, approverID := range req.Approvers {
		approvers = append(approvers, models.Approver{
			UserID:    approverID,
			Status:    models.StatusPending,
			Timestamp: time.Now(),
		})
	}

	// Tạo world state với approvers
	worldState := models.WorldState{
		ID:            primitive.NewObjectID(),
		TransactionID: req.TransactionID,
		FromAccount:   req.FromAccount,
		ToAccount:     req.ToAccount,
		Amount:        req.Amount,
		Status:        models.StatusPending,
		Approvers:     approvers,
		ApprovalCount: 0,
		LastUpdated:   time.Now(),
	}

	wsColl := db.GetCollection("world_state")
	_, err = wsColl.InsertOne(context.Background(), worldState)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, worldState)
}

type ApproveRequest struct {
	TransactionID string  `json:"transaction_id"`
	ApproverID    string  `json:"approver_id"`
	FromAccount   string  `json:"from_account"`
	ToAccount     string  `json:"to_account"`
	Amount        float64 `json:"amount"`
}

func ApproveTransaction(c *gin.Context) {
	var req ApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Tạo transaction approve
	tx := models.Transaction{
		ID:            primitive.NewObjectID(),
		TransactionID: req.TransactionID,
		FromAccount:   req.FromAccount,
		ToAccount:     req.ToAccount,
		Amount:        req.Amount,
		Type:          "APPROVE",
		Status:        models.StatusApproved,
		ApproverID:    req.ApproverID,
		Timestamp:     time.Now(),
	}

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

	// Cập nhật trạng thái của approver
	update := bson.M{
		"$set": bson.M{
			"last_updated":                time.Now(),
			"approvers.$[elem].status":    models.StatusApproved,
			"approvers.$[elem].timestamp": time.Now(),
		},
		"$inc": bson.M{"approval_count": 1},
	}

	arrayFilters := options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"elem.user_id": tx.ApproverID, "elem.status": models.StatusPending},
		},
	}

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After).
		SetArrayFilters(arrayFilters)

	var worldState models.WorldState
	err = wsColl.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&worldState)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Cập nhật trạng thái world state dựa trên số lượng approval
	newStatus := models.StatusPartiallyApproved
	if worldState.ApprovalCount >= config.ApprovalThreshold {
		newStatus = models.StatusApproved

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

		var supplierResp SupplierResponse
		if err := json.NewDecoder(resp.Body).Decode(&supplierResp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if supplierResp.Status == "SUCCESS" {
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

			// Cập nhật world state với supplier ref
			update = bson.M{
				"$set": bson.M{
					"status":       models.StatusExecuted,
					"supplier_ref": supplierResp.SupplierRef,
					"last_updated": time.Now(),
				},
			}
		} else {
			// Supplier failed
			newStatus = models.StatusApprovedPendingExec
		}

		// Cập nhật trạng thái cuối cùng
		update = bson.M{
			"$set": bson.M{
				"status":       newStatus,
				"last_updated": time.Now(),
			},
		}
		err = wsColl.FindOneAndUpdate(context.Background(), filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&worldState)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, worldState)
}

func GetPendingApprovals(c *gin.Context) {
	// Lấy userID từ query param
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Tìm các world state có approver là userID và status là pending
	wsColl := db.GetCollection("world_state")
	filter := bson.M{
		"approvers": bson.M{
			"$elemMatch": bson.M{
				"user_id": userID,
				"status":  models.StatusPending,
			},
		},
		"status": bson.M{
			"$in": []string{models.StatusPending, models.StatusPartiallyApproved},
		},
	}

	cursor, err := wsColl.Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	var pendingTransactions []models.WorldState
	if err = cursor.All(context.Background(), &pendingTransactions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pendingTransactions)
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
