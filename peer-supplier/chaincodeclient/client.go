package chaincodeclient

import (
	"context"
	"fmt"
	"log"

	pb "peer-supplier/share"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ChaincodeClient struct {
	conn   *grpc.ClientConn
	client pb.SmartContractServiceClient
}

func NewClient() *ChaincodeClient {
	conn, err := grpc.Dial("ms-scf-chaincode:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to chaincode service: %v", err)
	}

	return &ChaincodeClient{
		conn:   conn,
		client: pb.NewSmartContractServiceClient(conn),
	}
}

func (c *ChaincodeClient) Close() error {
	return c.conn.Close()
}

// Contract operations
func (c *ChaincodeClient) InvokeCreateContract(anchorID string, suppliers []string, totalAmount float64, fileHash string) (*pb.ContractResponse, error) {
	req := &pb.CreateContractRequest{
		AnchorId:    anchorID,
		Suppliers:   suppliers,
		TotalAmount: totalAmount,
		FileHash:    fileHash,
	}

	resp, err := c.client.CreateContract(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract: %w", err)
	}

	return resp, nil
}

func (c *ChaincodeClient) InvokeApproveContract(contractID, supplierID string) (*pb.ContractResponse, error) {
	req := &pb.ApproveContractRequest{
		ContractId: contractID,
		SupplierId: supplierID,
	}

	resp, err := c.client.ApproveContract(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to approve contract: %w", err)
	}

	return resp, nil
}

func (c *ChaincodeClient) InvokeFinalizeContract(contractID string) (*pb.ContractResponse, error) {
	req := &pb.FinalizeContractRequest{
		ContractId: contractID,
	}

	resp, err := c.client.FinalizeContract(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to finalize contract: %w", err)
	}

	return resp, nil
}

// Token operations
func (c *ChaincodeClient) InvokeIssueToken(contractID, issuer string, totalSupply float64) (*pb.TokenResponse, error) {
	req := &pb.IssueTokenRequest{
		ContractId:  contractID,
		Issuer:      issuer,
		TotalSupply: totalSupply,
	}

	resp, err := c.client.IssueToken(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to issue token: %w", err)
	}

	return resp, nil
}

func (c *ChaincodeClient) InvokeTransferToken(tokenID, from, to string, amount float64) (*pb.TokenResponse, error) {
	req := &pb.TransferTokenRequest{
		TokenId: tokenID,
		From:    from,
		To:      to,
		Amount:  amount,
	}

	resp, err := c.client.TransferToken(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer token: %w", err)
	}

	return resp, nil
}

func (c *ChaincodeClient) InvokeSettleToken(tokenID, supplierID, bankID string) (*pb.TokenResponse, error) {
	req := &pb.SettleTokenRequest{
		TokenId:    tokenID,
		SupplierId: supplierID,
		BankId:     bankID,
	}

	resp, err := c.client.SettleToken(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to settle token: %w", err)
	}

	return resp, nil
}
