package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"share"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to gRPC server
	conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := share.NewSmartContractServiceClient(conn)

	// Test CreateContract
	fmt.Println("Testing CreateContract...")

	req := &share.CreateContractRequest{
		AnchorId:    "ANCHOR001",
		Suppliers:   []string{"SUP001", "SUP002"},
		TotalAmount: 50000.0,
		FileHash:    "file_hash_123",
	}

	resp, err := client.CreateContract(context.Background(), req)
	if err != nil {
		log.Fatalf("CreateContract failed: %v", err)
	}

	// Wait a bit
	time.Sleep(1 * time.Second)

	// Test ApproveContract
	fmt.Println("\nTesting ApproveContract...")

	approveReq := &share.ApproveContractRequest{
		ContractId: resp.ContractId,
		SupplierId: "SUP001",
	}

	approveResp, err := client.ApproveContract(context.Background(), approveReq)
	if err != nil {
		log.Fatalf("ApproveContract failed: %v", err)
	}

	fmt.Println("\n✅ All tests completed successfully!")
}
