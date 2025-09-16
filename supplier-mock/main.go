package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ExecuteRequest struct {
	TransactionID string  `json:"transaction_id"`
	FromAccount   string  `json:"from_account"`
	ToAccount     string  `json:"to_account"`
	Amount        float64 `json:"amount"`
}

type ExecuteResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	r := gin.Default()

	r.POST("/supplier/execute", func(c *gin.Context) {
		var req ExecuteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Luôn trả về SUCCESS như yêu cầu
		response := ExecuteResponse{
			Status:    "SUCCESS",
			Timestamp: time.Now(),
		}

		c.JSON(http.StatusOK, response)
	})

	r.Run(":8082")
}
