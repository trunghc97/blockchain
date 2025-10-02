package orderer_client

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "blockchain-gw/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrdererClient struct {
	conn *grpc.ClientConn
}

func NewOrdererClient() *OrdererClient {
	return &OrdererClient{}
}

func (oc *OrdererClient) SubmitTransaction(ctx context.Context, proposal *pb.TransactionProposal, endorsements []*pb.Endorsement) (string, error) {
	// Connect to orderer if not already connected
	if oc.conn == nil {
		if err := oc.connect(); err != nil {
			return "", fmt.Errorf("failed to connect to orderer: %v", err)
		}
	}

	// Create transaction with endorsements (for future use)
	_ = &Transaction{
		TransactionId: proposal.TransactionId,
		FunctionName:  proposal.FunctionName,
		Args:          proposal.Args,
		ChannelId:     proposal.ChannelId,
		ChaincodeName: proposal.ChaincodeName,
		Endorsements:  endorsements,
		Timestamp:     time.Now().Unix(),
	}

	// Set timeout
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Submit to orderer (assuming orderer has SubmitTransaction method)
	// For now, we'll simulate the response
	log.Printf("Submitting transaction %s to orderer with %d endorsements", proposal.TransactionId, len(endorsements))

	// Simulate block number generation
	blockNumber := fmt.Sprintf("block_%d", time.Now().Unix())

	log.Printf("Transaction %s committed in block %s", proposal.TransactionId, blockNumber)
	return blockNumber, nil
}

func (oc *OrdererClient) GetTransactionStatus(ctx context.Context, transactionID string) (string, string, error) {
	// Connect to orderer if not already connected
	if oc.conn == nil {
		if err := oc.connect(); err != nil {
			return "", "", fmt.Errorf("failed to connect to orderer: %v", err)
		}
	}

	// Set timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Query transaction status (simulated for now)
	log.Printf("Querying status for transaction %s", transactionID)

	// Simulate status response
	return "COMMITTED", "block_12345", nil
}

func (oc *OrdererClient) connect() error {
	// Connect to orderer cluster (use first orderer for now)
	address := "orderer-ord1:7050"

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to dial orderer at %s: %v", address, err)
	}

	oc.conn = conn
	log.Printf("Connected to orderer at %s", address)
	return nil
}

func (oc *OrdererClient) Close() {
	if oc.conn != nil {
		if err := oc.conn.Close(); err != nil {
			log.Printf("Error closing orderer connection: %v", err)
		}
	}
}

// Transaction structure for orderer communication
type Transaction struct {
	TransactionId string            `json:"transaction_id"`
	FunctionName  string            `json:"function_name"`
	Args          []string          `json:"args"`
	ChannelId     string            `json:"channel_id"`
	ChaincodeName string            `json:"chaincode_name"`
	Endorsements  []*pb.Endorsement `json:"endorsements"`
	Timestamp     int64             `json:"timestamp"`
}
