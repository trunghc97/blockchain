package main

import (
	"log"
	"os"

	"shared/peer-base"
)

func main() {
	// Set default environment variables for anchor peer
	if os.Getenv("PEER_NODE_TYPE") == "" {
		os.Setenv("PEER_NODE_TYPE", "anchor")
	}
	if os.Getenv("PEER_NODE_ID") == "" {
		os.Setenv("PEER_NODE_ID", "anchor-peer-1")
	}
	if os.Getenv("PEER_PORT") == "" {
		os.Setenv("PEER_PORT", "8084")
	}
	if os.Getenv("PEER_GRPC_PORT") == "" {
		os.Setenv("PEER_GRPC_PORT", "9094")
	}
	if os.Getenv("ORDERER_ADDR") == "" {
		os.Setenv("ORDERER_ADDR", "orderer-cluster-ord1:7050")
	}

	log.Println("Starting Anchor Peer Node...")
	peer_base.Main()
}