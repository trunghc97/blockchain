package endorser

import (
	"context"
	"fmt"
	"log"
	"time"

	peerpb "blockchain-gw/peer-anchor-proto"
	pb "blockchain-gw/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type EndorserClient struct {
	peerConnections map[string]*grpc.ClientConn
}

func NewEndorserClient() *EndorserClient {
	return &EndorserClient{
		peerConnections: make(map[string]*grpc.ClientConn),
	}
}

func (ec *EndorserClient) EvaluateProposal(ctx context.Context, peerID string, proposal *pb.TransactionProposal) (*pb.Endorsement, error) {
	// Get or create connection to peer
	conn, err := ec.getConnection(peerID)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to peer %s: %v", peerID, err)
	}

	// Create client
	client := peerpb.NewPeerEndorsementClient(conn)

	// Create proposal request
	req := &peerpb.ProposalRequest{
		TransactionId: proposal.TransactionId,
		FunctionName:  proposal.FunctionName,
		Args:          proposal.Args,
		ChannelId:     proposal.ChannelId,
		ChaincodeName: proposal.ChaincodeName,
	}

	// Set timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Call peer endorsement service
	resp, err := client.EvaluateProposal(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("peer %s evaluation failed: %v", peerID, err)
	}

	// Convert to endorsement
	endorsement := &pb.Endorsement{
		PeerId:          peerID,
		Signature:       resp.Signature,
		RwSet:           convertReadWriteSet(resp.RwSet),
		ResponseStatus:  resp.Status,
		ResponseMessage: resp.Message,
	}

	log.Printf("Received endorsement from peer %s: status=%d", peerID, resp.Status)
	return endorsement, nil
}

func convertReadWriteSet(peerRwSet *peerpb.ReadWriteSet) *pb.ReadWriteSet {
	if peerRwSet == nil {
		return nil
	}

	return &pb.ReadWriteSet{
		ReadKeys: peerRwSet.ReadKeys,
		WriteSet: peerRwSet.WriteSet,
	}
}

func (ec *EndorserClient) getConnection(peerID string) (*grpc.ClientConn, error) {
	// Check if connection already exists
	if conn, exists := ec.peerConnections[peerID]; exists {
		return conn, nil
	}

	// Determine peer address based on peer ID
	var address string
	switch peerID {
	case "peer-anchor":
		address = "peer-anchor:9094"
	case "peer-main-bank":
		address = "peer-main-bank:9095"
	case "peer-supplier":
		address = "peer-supplier:9093"
	default:
		return nil, fmt.Errorf("unknown peer ID: %s", peerID)
	}

	// Create new connection
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial peer %s at %s: %v", peerID, address, err)
	}

	// Store connection
	ec.peerConnections[peerID] = conn
	log.Printf("Connected to peer %s at %s", peerID, address)

	return conn, nil
}

func (ec *EndorserClient) Close() {
	for peerID, conn := range ec.peerConnections {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection to peer %s: %v", peerID, err)
		}
	}
}
