package main

import (
	"ms-blockchain/blockchain"
	"ms-blockchain/db"
	"ms-blockchain/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Kết nối MongoDB
	db.Connect()

	// Khởi động block builder
	blockchain.StartBlockBuilder()

	// Khởi tạo Gin router
	r := gin.Default()

	// Định nghĩa các routes
	r.POST("/tx/create", handlers.CreateTransaction)
	r.POST("/tx/approve", handlers.ApproveTransaction)
	r.GET("/tx/status/:id", handlers.GetTransactionStatus)

	// Chạy server
	r.Run(":8081")
}
