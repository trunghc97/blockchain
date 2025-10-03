package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	"peer-supplier/blockstream"
	"peer-supplier/endorsement"
	"peer-supplier/handlers"
	pb "peer-supplier/proto"
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
		mongoURI = "mongodb://root:example@mongo-shared:27017/blockchain_private?authSource=admin"
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

	// Get private database for blockchain operations
	databaseName := "blockchain_private"
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		databaseName = dbName
	}
	db := client.Database(databaseName)

	// Get public database for user operations
	// publicDb := client.Database("blockchain_public")

	log.Printf("Connected to MongoDB databases: private=%s, public=blockchain_public", databaseName)

	// Initialize handlers
	h := handlers.NewHandler(db)

	// Initialize block streaming service (receive blocks from Orderer)
	blockStreamService := blockstream.NewBlockStreamService(db)
	ctx2 := context.Background()
	blockStreamService.StartBlockStreaming(ctx2)

	// Setup routes
	router := mux.NewRouter()

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "ok", "peer": "supplier", "message": "Peer is running"}`)
	}).Methods("GET")

	// Contract routes - Only GET endpoints for querying
	router.HandleFunc("/contract/list", h.GetContracts).Methods("GET")
	router.HandleFunc("/contract/{id}", h.GetContract).Methods("GET")
	router.HandleFunc("/contract/{id}/ledger", h.GetContractLedger).Methods("GET")

	// Token routes - Only GET endpoints for querying
	router.HandleFunc("/token/{id}", h.GetToken).Methods("GET")
	router.HandleFunc("/token/issued/{bankId}", h.GetTokensIssuedByBank).Methods("GET")

	// User/Balance routes - Only GET endpoints for querying
	router.HandleFunc("/users", h.GetSuppliers).Methods("GET")
	router.HandleFunc("/suppliers", h.GetSuppliers).Methods("GET")
	router.HandleFunc("/tokens", h.GetAllTokens).Methods("GET")
	router.HandleFunc("/balances/account/{accountId}", h.GetBalancesByAccount).Methods("GET")
	router.HandleFunc("/balances/token/{tokenId}", h.GetBalancesByToken).Methods("GET")

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

	// Start gRPC server for endorsement
	go func() {
		lis, err := net.Listen("tcp", ":9093")
		if err != nil {
			log.Fatalf("failed to listen on gRPC port: %v", err)
		}

		grpcServer := grpc.NewServer()
		endorsementService := endorsement.NewEndorsementService("peer-supplier")
		pb.RegisterPeerEndorsementServer(grpcServer, endorsementService)

		log.Printf("Supplier Peer gRPC server running on port 9093")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	log.Printf("Supplier Peer HTTP server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsHandler(router)))
}
