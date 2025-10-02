package gateway

import (
	"context"
	"fmt"
	"log"
	"sync"

	"blockchain-gw/endorser"
	orderer_client "blockchain-gw/orderer-client"
	pb "blockchain-gw/proto"
)

type BlockchainGateway struct {
	pb.UnimplementedBlockchainGatewayServer
	endorserClient *endorser.EndorserClient
	ordererClient  *orderer_client.OrdererClient
	policy         *EndorsementPolicy
}

type EndorsementPolicy struct {
	RequiredPeers   []string
	MinEndorsements int
}

func NewBlockchainGateway() *BlockchainGateway {
	endorserClient := endorser.NewEndorserClient()
	ordererClient := orderer_client.NewOrdererClient()

	// Default endorsement policy: require Anchor + Bank
	policy := &EndorsementPolicy{
		RequiredPeers:   []string{"peer-anchor", "peer-main-bank"},
		MinEndorsements: 2,
	}

	return &BlockchainGateway{
		endorserClient: endorserClient,
		ordererClient:  ordererClient,
		policy:         policy,
	}
}

func (bg *BlockchainGateway) SubmitTransaction(ctx context.Context, req *pb.TransactionProposal) (*pb.TransactionResponse, error) {
	log.Printf("Received transaction proposal: %s", req.TransactionId)

	// Step 1: Send proposal to all required peers for endorsement
	endorsements, err := bg.collectEndorsements(ctx, req)
	if err != nil {
		return &pb.TransactionResponse{
			TransactionId: req.TransactionId,
			Success:       false,
			Message:       fmt.Sprintf("Failed to collect endorsements: %v", err),
		}, nil
	}

	// Step 2: Check endorsement policy
	if !bg.checkEndorsementPolicy(endorsements) {
		return &pb.TransactionResponse{
			TransactionId: req.TransactionId,
			Success:       false,
			Message:       "Endorsement policy not satisfied",
		}, nil
	}

	// Step 3: Submit to orderer
	blockNumber, err := bg.ordererClient.SubmitTransaction(ctx, req, endorsements)
	if err != nil {
		return &pb.TransactionResponse{
			TransactionId: req.TransactionId,
			Success:       false,
			Message:       fmt.Sprintf("Failed to submit to orderer: %v", err),
		}, nil
	}

	return &pb.TransactionResponse{
		TransactionId: req.TransactionId,
		Success:       true,
		Message:       "Transaction committed successfully",
		BlockNumber:   blockNumber,
		Endorsements:  endorsements,
	}, nil
}

func (bg *BlockchainGateway) GetTransactionStatus(ctx context.Context, req *pb.TransactionStatusRequest) (*pb.TransactionStatusResponse, error) {
	// Query orderer for transaction status
	status, blockNumber, err := bg.ordererClient.GetTransactionStatus(ctx, req.TransactionId)
	if err != nil {
		return &pb.TransactionStatusResponse{
			TransactionId: req.TransactionId,
			Status:        "FAILED",
			ErrorMessage:  err.Error(),
		}, nil
	}

	return &pb.TransactionStatusResponse{
		TransactionId: req.TransactionId,
		Status:        status,
		BlockNumber:   blockNumber,
	}, nil
}

func (bg *BlockchainGateway) collectEndorsements(ctx context.Context, proposal *pb.TransactionProposal) ([]*pb.Endorsement, error) {
	var wg sync.WaitGroup
	endorsements := make([]*pb.Endorsement, 0)
	endorsementChan := make(chan *pb.Endorsement, len(bg.policy.RequiredPeers))
	errorChan := make(chan error, len(bg.policy.RequiredPeers))

	// Send proposal to all required peers concurrently
	for _, peerID := range bg.policy.RequiredPeers {
		wg.Add(1)
		go func(peer string) {
			defer wg.Done()

			endorsement, err := bg.endorserClient.EvaluateProposal(ctx, peer, proposal)
			if err != nil {
				errorChan <- fmt.Errorf("peer %s: %v", peer, err)
				return
			}

			endorsementChan <- endorsement
		}(peerID)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(endorsementChan)
	close(errorChan)

	// Collect endorsements
	for endorsement := range endorsementChan {
		endorsements = append(endorsements, endorsement)
	}

	// Check for errors
	if len(errorChan) > 0 {
		var errors []error
		for err := range errorChan {
			errors = append(errors, err)
		}
		return nil, fmt.Errorf("endorsement errors: %v", errors)
	}

	return endorsements, nil
}

func (bg *BlockchainGateway) checkEndorsementPolicy(endorsements []*pb.Endorsement) bool {
	if len(endorsements) < bg.policy.MinEndorsements {
		log.Printf("Insufficient endorsements: got %d, required %d", len(endorsements), bg.policy.MinEndorsements)
		return false
	}

	// Check if all required peers have endorsed
	endorsedPeers := make(map[string]bool)
	for _, endorsement := range endorsements {
		if endorsement.ResponseStatus == 200 { // Success status
			endorsedPeers[endorsement.PeerId] = true
		}
	}

	for _, requiredPeer := range bg.policy.RequiredPeers {
		if !endorsedPeers[requiredPeer] {
			log.Printf("Required peer %s did not endorse", requiredPeer)
			return false
		}
	}

	return true
}
