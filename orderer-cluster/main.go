package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"orderer-cluster/block"
	"orderer-cluster/grpc"
	"orderer-cluster/mempool"
	"orderer-cluster/pbft"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Get configuration from environment
	nodeID := getEnv("PBFT_NODE_ID", "ord1")
	port := getEnv("ORDERER_PORT", "7050")
	mongoURI := getEnv("MONGO_URI", "mongodb://root:example@mongo-shared:27017/blockchain?authSource=admin")
	fStr := getEnv("PBFT_F", "1")

	f, err := strconv.Atoi(fStr)
	if err != nil {
		log.Fatalf("Invalid PBFT_F value: %v", err)
	}

	// Load private key
	privateKey, err := loadPrivateKey(fmt.Sprintf("secrets/%s/private.pem", nodeID))
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	// Ping database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	db := client.Database("blockchain_public")
	log.Printf("Connected to MongoDB database: blockchain_public")

	// Initialize components
	mempool := mempool.NewMempool()
	blockBuilder := block.NewBlockBuilder(db)

	// Configure orderer cluster (3 nodes: ord1, ord2, ord3)
	orderers := map[string]*pbft.OrdererInfo{
		"ord1": {
			ID:      "ord1",
			Address: "orderer-ord1:7050",
		},
		"ord2": {
			ID:      "ord2",
			Address: "orderer-ord2:7060",
		},
		"ord3": {
			ID:      "ord3",
			Address: "orderer-ord3:7070",
		},
	}

	// Load public keys for all orderers
	for id, orderer := range orderers {
		pubKey, err := loadPublicKey(fmt.Sprintf("secrets/%s/public.pem", id))
		if err != nil {
			log.Fatalf("Failed to load public key for %s: %v", id, err)
		}
		orderer.PublicKey = pubKey
	}

	// Create PBFT node
	pbftNode := pbft.NewPBFTNode(nodeID, f, privateKey, orderers, mempool, blockBuilder, db)

	// Create gRPC server
	ordererServer := grpc.NewOrdererServer(pbftNode, blockBuilder)

	// Start PBFT consensus
	go pbftNode.Start(context.Background())

	// Start gRPC server
	log.Printf("Starting PBFT Orderer %s on port %s", nodeID, port)
	if err := grpc.StartServer(port, ordererServer); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func loadPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	pemData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKeyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	privateKey, ok := privateKeyInterface.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an ECDSA private key")
	}

	return privateKey, nil
}

func loadPublicKey(path string) (*ecdsa.PublicKey, error) {
	pemData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pubKey, ok := pubInterface.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an ECDSA public key")
	}

	return pubKey, nil
}
