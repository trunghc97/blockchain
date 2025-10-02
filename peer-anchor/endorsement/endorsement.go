package endorsement

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"

	pb "peer-anchor/proto"
)

type EndorsementService struct {
	pb.UnimplementedPeerEndorsementServer
	peerID string
}

func NewEndorsementService(peerID string) *EndorsementService {
	return &EndorsementService{
		peerID: peerID,
	}
}

func (es *EndorsementService) EvaluateProposal(ctx context.Context, req *pb.ProposalRequest) (*pb.ProposalResponse, error) {
	log.Printf("Evaluating proposal for transaction %s", req.TransactionId)

	// Step 1: Call smart contract engine to execute logic
	var rwSet *pb.ReadWriteSet
	var status int32 = 200
	var message string = "Success"

	switch req.FunctionName {
	case "CreateContract":
		rwSet, status, message = es.executeCreateContract(ctx, req)
	case "ApproveContract":
		rwSet, status, message = es.executeApproveContract(ctx, req)
	case "FinalizeContract":
		rwSet, status, message = es.executeFinalizeContract(ctx, req)
	case "IssueToken":
		rwSet, status, message = es.executeIssueToken(ctx, req)
	case "TransferToken":
		rwSet, status, message = es.executeTransferToken(ctx, req)
	case "SettleToken":
		rwSet, status, message = es.executeSettleToken(ctx, req)
	default:
		status = 400
		message = fmt.Sprintf("Unknown function: %s", req.FunctionName)
		rwSet = &pb.ReadWriteSet{}
	}

	// Step 2: Generate signature for RW set
	signature := es.generateSignature(rwSet)

	// Step 3: Return endorsement
	return &pb.ProposalResponse{
		TransactionId: req.TransactionId,
		Status:        status,
		Message:       message,
		RwSet:         rwSet,
		Signature:     signature,
		PeerId:        es.peerID,
	}, nil
}

func (es *EndorsementService) executeCreateContract(ctx context.Context, req *pb.ProposalRequest) (*pb.ReadWriteSet, int32, string) {
	// Execute contract creation logic directly
	contractID := fmt.Sprintf("contract_%s_%s", req.Args[0], req.Args[1])

	// Generate RW set
	rwSet := &pb.ReadWriteSet{
		ReadKeys: []string{},
		WriteSet: map[string]string{
			fmt.Sprintf("contract_%s", contractID): "CREATED",
		},
	}

	return rwSet, 200, "Contract created successfully"
}

func (es *EndorsementService) executeApproveContract(ctx context.Context, req *pb.ProposalRequest) (*pb.ReadWriteSet, int32, string) {
	// Execute contract approval logic directly
	contractID := req.Args[0]

	rwSet := &pb.ReadWriteSet{
		ReadKeys: []string{fmt.Sprintf("contract_%s", contractID)},
		WriteSet: map[string]string{
			fmt.Sprintf("contract_%s", contractID): "APPROVED",
		},
	}

	return rwSet, 200, "Contract approved successfully"
}

func (es *EndorsementService) executeFinalizeContract(ctx context.Context, req *pb.ProposalRequest) (*pb.ReadWriteSet, int32, string) {
	// Execute contract finalization logic directly
	contractID := req.Args[0]

	rwSet := &pb.ReadWriteSet{
		ReadKeys: []string{fmt.Sprintf("contract_%s", contractID)},
		WriteSet: map[string]string{
			fmt.Sprintf("contract_%s", contractID): "FINALIZED",
		},
	}

	return rwSet, 200, "Contract finalized successfully"
}

func (es *EndorsementService) executeIssueToken(ctx context.Context, req *pb.ProposalRequest) (*pb.ReadWriteSet, int32, string) {
	// Execute token issuance logic directly
	contractID := req.Args[0]
	tokenID := fmt.Sprintf("token_%s", contractID)

	rwSet := &pb.ReadWriteSet{
		ReadKeys: []string{fmt.Sprintf("contract_%s", contractID)},
		WriteSet: map[string]string{
			fmt.Sprintf("token_%s", tokenID): "ISSUED",
		},
	}

	return rwSet, 200, "Token issued successfully"
}

func (es *EndorsementService) executeTransferToken(ctx context.Context, req *pb.ProposalRequest) (*pb.ReadWriteSet, int32, string) {
	// Execute token transfer logic directly
	tokenID := req.Args[0]

	rwSet := &pb.ReadWriteSet{
		ReadKeys: []string{fmt.Sprintf("token_%s", tokenID)},
		WriteSet: map[string]string{
			fmt.Sprintf("token_%s", tokenID): "TRANSFERRED",
		},
	}

	return rwSet, 200, "Token transferred successfully"
}

func (es *EndorsementService) executeSettleToken(ctx context.Context, req *pb.ProposalRequest) (*pb.ReadWriteSet, int32, string) {
	// Execute token settlement logic directly
	tokenID := req.Args[0]

	rwSet := &pb.ReadWriteSet{
		ReadKeys: []string{fmt.Sprintf("token_%s", tokenID)},
		WriteSet: map[string]string{
			fmt.Sprintf("token_%s", tokenID): "SETTLED",
		},
	}

	return rwSet, 200, "Token settled successfully"
}

func (es *EndorsementService) generateSignature(rwSet *pb.ReadWriteSet) []byte {
	// Simple signature generation (in production, use proper cryptographic signing)
	data := fmt.Sprintf("%s_%s", es.peerID, es.serializeRWSet(rwSet))
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

func (es *EndorsementService) serializeRWSet(rwSet *pb.ReadWriteSet) string {
	result := ""
	for _, key := range rwSet.ReadKeys {
		result += fmt.Sprintf("R:%s,", key)
	}
	for key, value := range rwSet.WriteSet {
		result += fmt.Sprintf("W:%s=%s,", key, value)
	}
	return result
}
