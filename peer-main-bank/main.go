package main

import (
	"log"
	"os"

	"shared/peer-base"
)

func main() {
	// Set default environment variables for main-bank peer
	if os.Getenv("PEER_NODE_TYPE") == "" {
		os.Setenv("PEER_NODE_TYPE", "main-bank")
	}
	if os.Getenv("PEER_NODE_ID") == "" {
		os.Setenv("PEER_NODE_ID", "main-bank-peer-1")
	}
	if os.Getenv("PEER_PORT") == "" {
		os.Setenv("PEER_PORT", "8082") // Different port than original
	}
	if os.Getenv("PEER_GRPC_PORT") == "" {
		os.Setenv("PEER_GRPC_PORT", "9092")
	}
	if os.Getenv("ORDERER_ADDR") == "" {
		os.Setenv("ORDERER_ADDR", "orderer-cluster-ord1:7050")
	}

	log.Println("Starting Main Bank Peer Node...")
	peer_base.Main()
}