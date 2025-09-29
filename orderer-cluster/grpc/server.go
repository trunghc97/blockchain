package grpc

import (
	"context"
	"log"
	"net"

	"orderer-cluster/pbft"
	"orderer-cluster/proto"

	"google.golang.org/grpc"
)

// OrdererServer implements the OrdererService gRPC server
type OrdererServer struct {
	proto.UnimplementedOrdererServiceServer
	pbftNode     *pbft.PBFTNode
	blockBuilder pbft.BlockBuilder
}

// NewOrdererServer creates a new gRPC server
func NewOrdererServer(pbftNode *pbft.PBFTNode, blockBuilder pbft.BlockBuilder) *OrdererServer {
	return &OrdererServer{
		pbftNode:     pbftNode,
		blockBuilder: blockBuilder,
	}
}

// SubmitTx handles transaction submission
func (s *OrdererServer) SubmitTx(ctx context.Context, req *proto.SubmitTxRequest) (*proto.SubmitTxReply, error) {
	log.Printf("Received transaction %s from peer %s", req.Transaction.TransactionId, req.PeerId)

	// Submit transaction to PBFT node
	err := s.pbftNode.SubmitTransaction(req.Transaction)
	if err != nil {
		log.Printf("Failed to submit transaction: %v", err)
		return &proto.SubmitTxReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &proto.SubmitTxReply{
		Success:       true,
		TransactionId: req.Transaction.TransactionId,
		Message:       "Transaction submitted for consensus",
	}, nil
}

// StreamBlocks streams finalized blocks to peers
func (s *OrdererServer) StreamBlocks(req *proto.StreamBlocksRequest, stream proto.OrdererService_StreamBlocksServer) error {
	log.Printf("Peer %s started block streaming from height %d", req.PeerId, req.StartHeight)

	// Send existing blocks first
	startHeight := req.StartHeight
	if startHeight == 0 {
		startHeight = 1
	}

	for {
		// Get latest block height
		latestHeight := s.blockBuilder.GetLatestBlockHeight()

		// Send blocks from startHeight to latestHeight
		for height := startHeight; height <= latestHeight; height++ {
			block, err := s.blockBuilder.GetBlockByHeight(height)
			if err != nil {
				log.Printf("Failed to get block %d: %v", height, err)
				continue
			}

			if err := stream.Send(block); err != nil {
				log.Printf("Failed to send block %d: %v", height, err)
				return err
			}

			log.Printf("Sent block %d to peer %s", height, req.PeerId)
		}

		startHeight = latestHeight + 1

		// Wait for new blocks
		select {
		case block := <-s.pbftNode.GetBlocksChannel():
			if err := stream.Send(block); err != nil {
				log.Printf("Failed to send new block: %v", err)
				return err
			}
			log.Printf("Sent new block %d to peer %s", block.Height, req.PeerId)
			startHeight = block.Height + 1
		case <-stream.Context().Done():
			log.Printf("Peer %s disconnected from block stream", req.PeerId)
			return nil
		}
	}
}

// Consensus handles PBFT consensus messages between orderers
func (s *OrdererServer) Consensus(ctx context.Context, msg *proto.ConsensusMessage) (*proto.AckReply, error) {
	log.Printf("Received consensus message of type %s", msg.Type)

	// Forward to PBFT node
	s.pbftNode.ReceiveConsensusMessage(msg)

	return &proto.AckReply{
		Success: true,
		Message: "Consensus message received",
	}, nil
}

// StartServer starts the gRPC server
func StartServer(port string, server *OrdererServer) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	proto.RegisterOrdererServiceServer(s, server)

	log.Printf("gRPC server listening on port %s", port)
	return s.Serve(lis)
}
