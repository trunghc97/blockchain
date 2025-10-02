package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"google.golang.org/grpc"

	"peer-anchor/blockchain"
	"peer-anchor/config"
	"peer-anchor/db"
	"peer-anchor/endorsement"
	"peer-anchor/handlers"
	pb "peer-anchor/proto"
)

func main() {
	// Initialize MongoDB connection
	client, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

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
	router.HandleFunc("/contract/{id}/approve", handler.ApproveContract).Methods("POST")
	router.HandleFunc("/contract/{id}/approve-bank", handler.ApproveContractByBank).Methods("POST")
	router.HandleFunc("/contracts", handler.GetContracts).Methods("GET")
	router.HandleFunc("/contract/{id}", handler.GetContract).Methods("GET")
	router.HandleFunc("/contract/{id}/ledger", handler.GetContractLedger).Methods("GET")

	// Token endpoints
	router.HandleFunc("/token/{id}", handler.GetToken).Methods("GET")
	router.HandleFunc("/token/transfer", handler.TransferToken).Methods("POST")
	router.HandleFunc("/token/settle", handler.SettleToken).Methods("POST")
	router.HandleFunc("/token/issued/{bankId}", handler.GetTokensIssuedByBank).Methods("GET")
	router.HandleFunc("/tokens", handler.GetAllTokens).Methods("GET")
	router.HandleFunc("/balances/account/{accountId}", handler.GetBalancesByAccount).Methods("GET")
	router.HandleFunc("/balances/token/{tokenId}", handler.GetBalancesByToken).Methods("GET")

	// Supplier endpoints
	router.HandleFunc("/suppliers", handler.GetSuppliers).Methods("GET")

	// Utility endpoints
	router.HandleFunc("/blocks/hash/update", handler.UpdateBlockHashes).Methods("POST")

	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	// Start gRPC server for endorsement
	go func() {
		lis, err := net.Listen("tcp", ":9094")
		if err != nil {
			log.Fatalf("failed to listen on gRPC port: %v", err)
		}

		grpcServer := grpc.NewServer()
		endorsementService := endorsement.NewEndorsementService("peer-anchor")
		pb.RegisterPeerEndorsementServer(grpcServer, endorsementService)

		log.Printf("Anchor Peer gRPC server running on port 9094")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// Start HTTP server
	srv := &http.Server{
		Handler:      c.Handler(router),
		Addr:         ":8084",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Anchor Peer HTTP server running on port 8084")
	log.Fatal(srv.ListenAndServe())
}
