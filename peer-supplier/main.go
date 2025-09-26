package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"peer-supplier/handlers"
)

func main() {
	// Set default environment variables for supplier peer
	if os.Getenv("PEER_NODE_TYPE") == "" {
		os.Setenv("PEER_NODE_TYPE", "supplier")
	}
	if os.Getenv("PEER_NODE_ID") == "" {
		os.Setenv("PEER_NODE_ID", "supplier-peer-1")
	}
	if os.Getenv("PEER_PORT") == "" {
		os.Setenv("PEER_PORT", "8083")
	}

	port := os.Getenv("PEER_PORT")
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://root:example@mongo-supplier:27017/blockchain_supplier?authSource=admin"
	}

	log.Printf("Supplier Peer starting on port %s", port)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Disconnect(ctx)

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	// Get database
	databaseName := "blockchain_supplier"
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		databaseName = dbName
	}
	db := client.Database(databaseName)

	log.Printf("Connected to MongoDB database: %s", databaseName)

	// Initialize handlers
	h := handlers.NewHandler(db)

	// Setup routes
	router := mux.NewRouter()

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "ok", "peer": "supplier", "message": "Peer is running"}`)
	}).Methods("GET")

	// Contract routes
	router.HandleFunc("/contract/create", h.CreateContract).Methods("POST")
	router.HandleFunc("/contract/list", h.GetContracts).Methods("GET")
	router.HandleFunc("/contract/{id}", h.GetContract).Methods("GET")
	router.HandleFunc("/contract/{id}/approve", h.ApproveContract).Methods("POST")
	router.HandleFunc("/contract/{id}/approve-bank", h.ApproveContractByBank).Methods("POST")
	router.HandleFunc("/contract/{id}/ledger", h.GetContractLedger).Methods("GET")

	// Token routes
	router.HandleFunc("/token/{id}", h.GetToken).Methods("GET")
	router.HandleFunc("/token/transfer", h.TransferToken).Methods("POST")
	router.HandleFunc("/token/issued/{bankId}", h.GetTokensIssuedByBank).Methods("GET")
	router.HandleFunc("/token/settle", h.SettleToken).Methods("POST")

	// User/Balance routes
	router.HandleFunc("/users", h.GetSuppliers).Methods("GET")
	router.HandleFunc("/suppliers", h.GetSuppliers).Methods("GET")
	router.HandleFunc("/tokens", h.GetAllTokens).Methods("GET")
	router.HandleFunc("/balances/account/{accountId}", h.GetBalancesByAccount).Methods("GET")
	router.HandleFunc("/balances/token/{tokenId}", h.GetBalancesByToken).Methods("GET")

	// Admin routes
	router.HandleFunc("/admin/update-hashes", h.UpdateBlockHashes).Methods("POST")

	// CORS middleware
	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "*")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	log.Printf("Supplier Peer listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsHandler(router)))
}
