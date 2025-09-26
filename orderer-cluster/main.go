package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	pb "orderer-cluster/proto"
)

type OrdererNode struct {
	nodeID       string
	port         int
	peers        []string
	blockBuffer  chan *pb.Block
	mutex        sync.Mutex
	currentBlock *pb.Block
}

type OrdererService struct {
	pb.UnimplementedOrdererServiceServer
	orderer *OrdererNode
}

func NewOrdererNode(nodeID string, port int, peers []string) *OrdererNode {
	return &OrdererNode{
		nodeID:      nodeID,
		port:        port,
		peers:       peers,
		blockBuffer: make(chan *pb.Block, 100),
	}
}

func (o *OrdererNode) Start() {
	// Start block ordering goroutine
	go o.orderBlocks()

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", o.port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrdererServiceServer(grpcServer, &OrdererService{orderer: o})
	reflection.Register(grpcServer)

	log.Printf("Orderer %s listening on port %d", o.nodeID, o.port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (o *OrdererNode) orderBlocks() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case block := <-o.blockBuffer:
			o.processBlock(block)
		case <-ticker.C:
			o.createOrderedBlock()
		}
	}
}

func (o *OrdererNode) processBlock(block *pb.Block) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	log.Printf("Orderer %s processing block %d", o.nodeID, block.BlockNumber)

	// Validate block and add to ordering queue
	// In a real implementation, this would involve consensus among orderer nodes

	// Broadcast to other orderer nodes for consensus
	o.broadcastToPeers(block)
}

func (o *OrdererNode) createOrderedBlock() {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.currentBlock == nil {
		return
	}

	// Create ordered block with consensus
	orderedBlock := &pb.Block{
		BlockNumber: o.currentBlock.BlockNumber,
		Timestamp:   time.Now().Unix(),
		Transactions: o.currentBlock.Transactions,
		PreviousHash: o.currentBlock.PreviousHash,
		Hash:         o.calculateBlockHash(o.currentBlock),
		OrdererId:    o.nodeID,
	}

	log.Printf("Orderer %s created ordered block %d", o.nodeID, orderedBlock.BlockNumber)

	// Broadcast ordered block to peers
	o.broadcastOrderedBlock(orderedBlock)

	o.currentBlock = nil
}

func (o *OrdererNode) broadcastToPeers(block *pb.Block) {
	// Implementation for broadcasting to other orderer nodes
	log.Printf("Orderer %s broadcasting block to peers", o.nodeID)
}

func (o *OrdererNode) broadcastOrderedBlock(block *pb.Block) {
	// Implementation for broadcasting ordered block to connected peers
	log.Printf("Orderer %s broadcasting ordered block %d", o.nodeID, block.BlockNumber)
}

func (o *OrdererNode) calculateBlockHash(block *pb.Block) string {
	// Simple hash calculation - in real implementation use proper crypto
	return fmt.Sprintf("hash_%d_%s", block.BlockNumber, o.nodeID)
}

func (s *OrdererService) SubmitBlock(ctx context.Context, req *pb.SubmitBlockRequest) (*pb.SubmitBlockResponse, error) {
	block := req.Block
	log.Printf("Orderer %s received block %d from peer %s", s.orderer.nodeID, block.BlockNumber, req.PeerId)

	// Add to ordering queue
	s.orderer.blockBuffer <- block

	return &pb.SubmitBlockResponse{
		Success: true,
		Message: "Block submitted for ordering",
	}, nil
}

func (s *OrdererService) GetOrderedBlocks(ctx context.Context, req *pb.GetOrderedBlocksRequest) (*pb.GetOrderedBlocksResponse, error) {
	// Return ordered blocks for the requesting peer
	return &pb.GetOrderedBlocksResponse{
		Blocks: []*pb.Block{}, // Implement proper block retrieval
	}, nil
}

func main() {
	nodeID := os.Getenv("ORDERER_NODE_ID")
	if nodeID == "" {
		nodeID = "ord1"
	}

	portStr := os.Getenv("ORDERER_PORT")
	port := 7050 // default port
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	// Configure peers based on node ID
	var peers []string
	switch nodeID {
	case "ord1":
		peers = []string{"ord2:7060", "ord3:7070"}
	case "ord2":
		peers = []string{"ord1:7050", "ord3:7070"}
	case "ord3":
		peers = []string{"ord1:7050", "ord2:7060"}
	}

	orderer := NewOrdererNode(nodeID, port, peers)
	orderer.Start()
}
