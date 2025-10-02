package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"blockchain-gw/gateway"
	"blockchain-gw/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type HTTPRequest struct {
	TransactionId string   `json:"transaction_id"`
	FunctionName  string   `json:"function_name"`
	Args          []string `json:"args"`
	ChannelId     string   `json:"channel_id"`
	ChaincodeName string   `json:"chaincode_name"`
}

type HTTPResponse struct {
	TransactionId string `json:"transaction_id"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	BlockNumber   string `json:"block_number"`
}

func main() {
	// Start gRPC server
	go startGRPCServer()

	// Start HTTP server
	startHTTPServer()
}

func startGRPCServer() {
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	blockchainGW := gateway.NewBlockchainGateway()
	proto.RegisterBlockchainGatewayServer(s, blockchainGW)

	log.Println("Blockchain Gateway gRPC running on :9090")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func startHTTPServer() {
	http.HandleFunc("/submit-transaction", handleSubmitTransaction)
	http.HandleFunc("/health", handleHealth)

	log.Println("Blockchain Gateway HTTP running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleSubmitTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req HTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Convert HTTP request to gRPC request
	grpcReq := &proto.TransactionProposal{
		TransactionId: req.TransactionId,
		FunctionName:  req.FunctionName,
		Args:          req.Args,
		ChannelId:     req.ChannelId,
		ChaincodeName: req.ChaincodeName,
	}

	// Call gRPC service
	conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to gRPC server: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	client := proto.NewBlockchainGatewayClient(conn)
	resp, err := client.SubmitTransaction(r.Context(), grpcReq)
	if err != nil {
		log.Printf("gRPC call failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert gRPC response to HTTP response
	httpResp := HTTPResponse{
		TransactionId: resp.TransactionId,
		Success:       resp.Success,
		Message:       resp.Message,
		BlockNumber:   resp.BlockNumber,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(httpResp)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}
