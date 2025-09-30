package grpcclient

import (
	"context"
	"log"
	"time"

	"peer-supplier/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OrdererClient handles gRPC communication with orderer cluster
type OrdererClient struct {
	conn   *grpc.ClientConn
	client proto.OrdererServiceClient
}

// NewOrdererClient creates a new gRPC client to orderer
func NewOrdererClient(ordererAddr string) (*OrdererClient, error) {
	conn, err := grpc.Dial(ordererAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := proto.NewOrdererServiceClient(conn)

	return &OrdererClient{
		conn:   conn,
		client: client,
	}, nil
}

// Close closes the gRPC connection
func (oc *OrdererClient) Close() error {
	return oc.conn.Close()
}

// SubmitTransaction submits a transaction to the orderer cluster
func (oc *OrdererClient) SubmitTransaction(peerID string, tx *proto.Transaction) error {
	req := &proto.SubmitTxRequest{
		PeerId:      peerID,
		Transaction: tx,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := oc.client.SubmitTx(ctx, req)
	if err != nil {
		return err
	}

	if !resp.Success {
		log.Printf("Transaction submission failed: %s", resp.Message)
		return nil // Don't fail, just log
	}

	log.Printf("Transaction %s submitted successfully: %s", resp.TransactionId, resp.Message)
	return nil
}

// StreamBlocks starts streaming blocks from the orderer
func (oc *OrdererClient) StreamBlocks(peerID string, startHeight int64, blockHandler func(*proto.Block)) error {
	req := &proto.StreamBlocksRequest{
		PeerId:      peerID,
		StartHeight: startHeight,
	}

	stream, err := oc.client.StreamBlocks(context.Background(), req)
	if err != nil {
		return err
	}

	for {
		block, err := stream.Recv()
		if err != nil {
			return err
		}

		log.Printf("Received block %d from orderer", block.Height)
		blockHandler(block)
	}
}
