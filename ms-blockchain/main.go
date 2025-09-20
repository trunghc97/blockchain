package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"ms-blockchain/blockchain"
	"ms-blockchain/config"
	"ms-blockchain/db"
	"ms-blockchain/handlers"
)

func main() {
	// Initialize MongoDB connection
	client, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(nil)

	database := client.Database(config.DatabaseName)

	// Initialize handlers
	handler := handlers.NewHandler(database)

	// Initialize block builder
	blockBuilder := blockchain.NewBlockBuilder(database)
	blockBuilder.Start()

	// Initialize router
	router := mux.NewRouter()

	// Contract endpoints
	router.HandleFunc("/contract/create", handler.CreateContract).Methods("POST")
	router.HandleFunc("/contract/approve", handler.ApproveContract).Methods("POST")
	router.HandleFunc("/contract/list", handler.ListContracts).Methods("GET")

	// Ledger query endpoint
	router.HandleFunc("/ledger/query", handler.QueryLedger).Methods("GET")

	// User endpoints
	router.HandleFunc("/users", handler.GetUsers).Methods("GET")

	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	// Start server
	srv := &http.Server{
		Handler:      c.Handler(router),
		Addr:         ":8081",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Server is running on port 8081")
	log.Fatal(srv.ListenAndServe())
}
