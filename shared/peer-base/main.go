package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"shared/peer-base/blockchain"
	"shared/peer-base/config"
	"shared/peer-base/db"
	"shared/peer-base/handlers"
	pb "shared/proto"
)

type PeerNode struct {
	nodeID        string
	nodeType      string // "main-bank", "supplier", "anchor"
	port          int
	grpcPort      int
	ordererAddr   string
	db            *mongo.Database
	blockBuilder  *blockchain.BlockBuilder
	grpcClient    pb.OrdererServiceClient
	kafkaProducer *KafkaProducer
}

func NewPeerNode(nodeType string) *PeerNode {
	nodeID := os.Getenv("PEER_NODE_ID")
	if nodeID == "" {
		nodeID = fmt.Sprintf("%s-peer-1", nodeType)
	}

	portStr := os.Getenv("PEER_PORT")
	port := 8081 // default port
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	grpcPortStr := os.Getenv("PEER_GRPC_PORT")
	grpcPort := 9091 // default gRPC port
	if grpcPortStr != "" {
		if p, err := strconv.Atoi(grpcPortStr); err == nil {
			grpcPort = p
		}
	}

	ordererAddr := os.Getenv("ORDERER_ADDR")
	if ordererAddr == "" {
		ordererAddr = "orderer-cluster-ord1:7050"
	}

	return &PeerNode{
		nodeID:      nodeID,
		nodeType:    nodeType,
		port:        port,
		grpcPort:    grpcPort,
		ordererAddr: ordererAddr,
	}
}

func (p *PeerNode) Initialize() error {
	// Connect to MongoDB (peer-specific database)
	client, err := db.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Use peer-specific database name
	dbName := fmt.Sprintf("blockchain_%s", p.nodeType)
	p.db = client.Database(dbName)

	// Initialize handlers
	handler := handlers.NewHandler(p.db, p.nodeType, p)

	// Initialize block builder
	p.blockBuilder = blockchain.NewBlockBuilder(p.db)
	p.blockBuilder.Start()

	// Initialize Kafka producer
	if err := p.initKafkaProducer(); err != nil {
		log.Printf("Warning: Failed to initialize Kafka producer: %v", err)
	}

	// Connect to Orderer cluster
	if err := p.connectToOrderer(); err != nil {
		log.Printf("Warning: Failed to connect to orderer: %v", err)
	}

	// Start HTTP server
	router := p.setupRoutes(handler)

	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	// Start gRPC server
	if err := p.startGRPCServer(); err != nil {
		log.Printf("Warning: Failed to start gRPC server: %v", err)
	}

	// Start peer synchronization
	go p.syncWithPeers()

	// Start HTTP server
	srv := &http.Server{
		Handler:      c.Handler(router),
		Addr:         fmt.Sprintf(":%d", p.port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Peer %s (%s) listening on port %d (HTTP) and %d (gRPC)", p.nodeID, p.nodeType, p.port, p.grpcPort)
	log.Fatal(srv.ListenAndServe())

	return nil
}

func (p *PeerNode) connectToOrderer() error {
	conn, err := grpc.Dial(p.ordererAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	p.grpcClient = pb.NewOrdererServiceClient(conn)
	log.Printf("Connected to Orderer at %s", p.ordererAddr)
	return nil
}

func (p *PeerNode) setupRoutes(handler *handlers.Handler) *mux.Router {
	router := mux.NewRouter()

	// Contract endpoints (available on all peer types)
	router.HandleFunc("/contract/create", handler.CreateContract).Methods("POST")
	router.HandleFunc("/contract/{id}/approve", handler.ApproveContract).Methods("POST")
	router.HandleFunc("/contract/{id}/approve-bank", handler.ApproveContractByBank).Methods("POST")
	router.HandleFunc("/contract/list", handler.GetContracts).Methods("GET")
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

	// Peer-specific routes based on node type
	switch p.nodeType {
	case "main-bank":
		p.setupBankRoutes(router, handler)
	case "supplier":
		p.setupSupplierRoutes(router, handler)
	case "anchor":
		p.setupAnchorRoutes(router, handler)
	}

	return router
}

func (p *PeerNode) setupBankRoutes(router *mux.Router, handler *handlers.Handler) {
	// Bank-specific endpoints
	router.HandleFunc("/bank/dashboard", handler.GetBankDashboard).Methods("GET")
	router.HandleFunc("/bank/tokens/issued", handler.GetBankIssuedTokens).Methods("GET")
	router.HandleFunc("/bank/contracts/pending", handler.GetBankPendingContracts).Methods("GET")
}

func (p *PeerNode) setupSupplierRoutes(router *mux.Router, handler *handlers.Handler) {
	// Supplier-specific endpoints
	router.HandleFunc("/supplier/dashboard", handler.GetSupplierDashboard).Methods("GET")
	router.HandleFunc("/supplier/contracts", handler.GetSupplierContracts).Methods("GET")
	router.HandleFunc("/supplier/tokens", handler.GetSupplierTokens).Methods("GET")
}

func (p *PeerNode) setupAnchorRoutes(router *mux.Router, handler *handlers.Handler) {
	// Anchor-specific endpoints
	router.HandleFunc("/anchor/dashboard", handler.GetAnchorDashboard).Methods("GET")
	router.HandleFunc("/anchor/contracts", handler.GetAnchorContracts).Methods("GET")
	router.HandleFunc("/anchor/upload", handler.UploadContractFile).Methods("POST")
}

func (p *PeerNode) SubmitToOrderer(block *pb.Block) error {
	if p.grpcClient == nil {
		return fmt.Errorf("not connected to orderer")
	}

	req := &pb.SubmitBlockRequest{
		PeerId: p.nodeID,
		Block:  block,
	}

	resp, err := p.grpcClient.SubmitBlock(context.Background(), req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("orderer rejected block: %s", resp.Message)
	}

	log.Printf("Block %d submitted to orderer successfully", block.BlockNumber)
	return nil
}

// PublishEventToKafka publishes blockchain events to Kafka
func (p *PeerNode) PublishEventToKafka(event map[string]interface{}, channel string) error {
	return p.publishEventToKafka(event, channel)
}

// PublishTransactionToKafka publishes transactions to Kafka
func (p *PeerNode) PublishTransactionToKafka(transaction *pb.Transaction, channel string) error {
	return p.publishTransactionToKafka(transaction, channel)
}

func main() {
	nodeType := os.Getenv("PEER_NODE_TYPE")
	if nodeType == "" {
		nodeType = "main-bank" // default
	}

	peer := NewPeerNode(nodeType)
	if err := peer.Initialize(); err != nil {
		log.Fatal(err)
	}
}