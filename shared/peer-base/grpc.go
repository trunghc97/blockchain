package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "shared/proto"
)

type PeerGRPCServer struct {
	pb.UnimplementedPeerServiceServer
	peerNode *PeerNode
}

type PeerGRPCClient struct {
	conn   *grpc.ClientConn
	client pb.PeerServiceClient
}

func NewPeerGRPCServer(peerNode *PeerNode) *PeerGRPCServer {
	return &PeerGRPCServer{peerNode: peerNode}
}

func (s *PeerGRPCServer) SubmitTransaction(ctx context.Context, req *pb.SubmitTransactionRequest) (*pb.SubmitTransactionResponse, error) {
	log.Printf("Received transaction from peer %s", req.PeerId)

	// Validate and process transaction
	// This would involve business logic validation

	return &pb.SubmitTransactionResponse{
		Success: true,
		Message: "Transaction submitted successfully",
		TransactionId: req.Transaction.TransactionId,
	}, nil
}

func (s *PeerGRPCServer) GetBlocks(ctx context.Context, req *pb.GetBlocksRequest) (*pb.GetBlocksResponse, error) {
	log.Printf("Peer %s requesting blocks from %d to %d", req.PeerId, req.StartBlock, req.EndBlock)

	// Query blocks from database
	// Convert to protobuf format

	return &pb.GetBlocksResponse{
		Blocks: []*pb.Block{}, // Implement block retrieval
	}, nil
}

func (s *PeerGRPCServer) SyncWorldState(ctx context.Context, req *pb.SyncWorldStateRequest) (*pb.SyncWorldStateResponse, error) {
	log.Printf("Peer %s requesting world state sync", req.PeerId)

	// Get new blocks since last sync
	// Return world state changes

	return &pb.SyncWorldStateResponse{
		Success: true,
		Message: "World state synced successfully",
		NewBlocks: []*pb.Block{},
	}, nil
}

func (s *PeerGRPCServer) ValidateTransaction(ctx context.Context, req *pb.ValidateTransactionRequest) (*pb.ValidateTransactionResponse, error) {
	log.Printf("Validating transaction from peer %s", req.PeerId)

	// Validate transaction business rules
	// Check balances, permissions, etc.

	return &pb.ValidateTransactionResponse{
		Valid: true,
		Message: "Transaction is valid",
	}, nil
}

func NewPeerGRPCClient(address string) (*PeerGRPCClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewPeerServiceClient(conn)
	return &PeerGRPCClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *PeerGRPCClient) Close() error {
	return c.conn.Close()
}

func (c *PeerGRPCClient) SubmitTransaction(peerId string, transaction *pb.Transaction) error {
	req := &pb.SubmitTransactionRequest{
		PeerId:      peerId,
		Transaction: transaction,
	}

	resp, err := c.client.SubmitTransaction(context.Background(), req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("transaction submission failed: %s", resp.Message)
	}

	log.Printf("Transaction submitted successfully: %s", resp.TransactionId)
	return nil
}

func (c *PeerGRPCClient) SyncWithPeer(peerId string) error {
	req := &pb.SyncWorldStateRequest{
		PeerId:      peerId,
		LastSyncHash: "", // Implement proper sync hash
	}

	resp, err := c.client.SyncWorldState(context.Background(), req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("sync failed: %s", resp.Message)
	}

	log.Printf("Synced %d new blocks from peer", len(resp.NewBlocks))
	return nil
}

func (p *PeerNode) startGRPCServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", p.grpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port %d: %v", p.grpcPort, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPeerServiceServer(grpcServer, NewPeerGRPCServer(p))

	log.Printf("Peer %s gRPC server listening on port %d", p.nodeID, p.grpcPort)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	return nil
}

func (p *PeerNode) syncWithPeers() {
	// Define peer addresses based on node type
	var peerAddresses []string

	switch p.nodeType {
	case "main-bank":
		peerAddresses = []string{
			"peer-supplier:9093",
			"peer-anchor:9094",
		}
	case "supplier":
		peerAddresses = []string{
			"peer-main-bank:9092",
			"peer-anchor:9094",
		}
	case "anchor":
		peerAddresses = []string{
			"peer-main-bank:9092",
			"peer-supplier:9093",
		}
	}

	// Sync with other peers periodically
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for _, addr := range peerAddresses {
			client, err := NewPeerGRPCClient(addr)
			if err != nil {
				log.Printf("Failed to connect to peer %s: %v", addr, err)
				continue
			}

			err = client.SyncWithPeer(p.nodeID)
			if err != nil {
				log.Printf("Failed to sync with peer %s: %v", addr, err)
			}

			client.Close()
		}
	}
}
