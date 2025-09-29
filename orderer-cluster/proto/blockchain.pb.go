// Simple protobuf definitions for blockchain
package proto

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Block represents a blockchain block with PBFT signatures
type Block struct {
	Height       int64             `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	Timestamp    *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Transactions []*Transaction     `protobuf:"bytes,3,rep,name=transactions,proto3" json:"transactions,omitempty"`
	PreviousHash string            `protobuf:"bytes,4,opt,name=previous_hash,json=previousHash,proto3" json:"previous_hash,omitempty"`
	Hash         string            `protobuf:"bytes,5,opt,name=hash,proto3" json:"hash,omitempty"`
	MerkleRoot   string            `protobuf:"bytes,6,opt,name=merkle_root,json=merkleRoot,proto3" json:"merkle_root,omitempty"`
	Signatures   []*BlockSignature `protobuf:"bytes,7,rep,name=signatures,proto3" json:"signatures,omitempty"`
}

// BlockSignature represents a PBFT signature from an orderer
type BlockSignature struct {
	OrdererId string `protobuf:"bytes,1,opt,name=orderer_id,json=ordererId,proto3" json:"orderer_id,omitempty"`
	Signature []byte `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
	PublicKey string `protobuf:"bytes,3,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
}

// Transaction represents a blockchain transaction
type Transaction struct {
	TransactionId   string                `protobuf:"bytes,1,opt,name=transaction_id,json=transactionId,proto3" json:"transaction_id,omitempty"`
	TransactionType string                `protobuf:"bytes,2,opt,name=transaction_type,json=transactionType,proto3" json:"transaction_type,omitempty"`
	ContractId      string                `protobuf:"bytes,3,opt,name=contract_id,json=contractId,proto3" json:"contract_id,omitempty"`
	TokenId         string                `protobuf:"bytes,4,opt,name=token_id,json=tokenId,proto3" json:"token_id,omitempty"`
	SenderId        string                `protobuf:"bytes,5,opt,name=sender_id,json=senderId,proto3" json:"sender_id,omitempty"`
	ReceiverId      string                `protobuf:"bytes,6,opt,name=receiver_id,json=receiverId,proto3" json:"receiver_id,omitempty"`
	Amount          float64               `protobuf:"fixed64,7,opt,name=amount,proto3" json:"amount,omitempty"`
	Payload         string                `protobuf:"bytes,8,opt,name=payload,proto3" json:"payload,omitempty"`
	Timestamp       *timestamppb.Timestamp `protobuf:"bytes,9,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

// Consensus message types for PBFT
type PrePrepareMessage struct {
	View           string `protobuf:"bytes,1,opt,name=view,proto3" json:"view,omitempty"`
	SequenceNumber string `protobuf:"bytes,2,opt,name=sequence_number,json=sequenceNumber,proto3" json:"sequence_number,omitempty"`
	Block          *Block `protobuf:"bytes,3,opt,name=block,proto3" json:"block,omitempty"`
	Digest         string `protobuf:"bytes,4,opt,name=digest,proto3" json:"digest,omitempty"`
}

type PrepareMessage struct {
	View           string `protobuf:"bytes,1,opt,name=view,proto3" json:"view,omitempty"`
	SequenceNumber string `protobuf:"bytes,2,opt,name=sequence_number,json=sequenceNumber,proto3" json:"sequence_number,omitempty"`
	Digest         string `protobuf:"bytes,3,opt,name=digest,proto3" json:"digest,omitempty"`
	OrdererId      string `protobuf:"bytes,4,opt,name=orderer_id,json=ordererId,proto3" json:"orderer_id,omitempty"`
}

type CommitMessage struct {
	View           string `protobuf:"bytes,1,opt,name=view,proto3" json:"view,omitempty"`
	SequenceNumber string `protobuf:"bytes,2,opt,name=sequence_number,json=sequenceNumber,proto3" json:"sequence_number,omitempty"`
	Digest         string `protobuf:"bytes,3,opt,name=digest,proto3" json:"digest,omitempty"`
	OrdererId      string `protobuf:"bytes,4,opt,name=orderer_id,json=ordererId,proto3" json:"orderer_id,omitempty"`
}

type ConsensusMessage struct {
	Type string `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	// Types that are assignable to Message:
	PrePrepare *PrePrepareMessage `protobuf:"bytes,2,opt,name=pre_prepare,json=prePrepare,proto3,oneof" json:"pre_prepare,omitempty"`
	Prepare    *PrepareMessage    `protobuf:"bytes,3,opt,name=prepare,proto3,oneof" json:"prepare,omitempty"`
	Commit     *CommitMessage     `protobuf:"bytes,4,opt,name=commit,proto3,oneof" json:"commit,omitempty"`
}

// Service messages
type SubmitTxRequest struct {
	PeerId      string       `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	Transaction *Transaction `protobuf:"bytes,2,opt,name=transaction,proto3" json:"transaction,omitempty"`
}

type SubmitTxReply struct {
	Success       bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	TransactionId string `protobuf:"bytes,2,opt,name=transaction_id,json=transactionId,proto3" json:"transaction_id,omitempty"`
	Message       string `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
}

type StreamBlocksRequest struct {
	PeerId     string `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	StartHeight int64  `protobuf:"varint,2,opt,name=start_height,json=startHeight,proto3" json:"start_height,omitempty"`
}

type AckReply struct {
	Success bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

// Helper methods for protobuf compatibility
func (x *Block) GetHeight() int64 {
	if x != nil {
		return x.Height
	}
	return 0
}

func (x *Block) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

func (x *Block) GetTransactions() []*Transaction {
	if x != nil {
		return x.Transactions
	}
	return nil
}

func (x *Block) GetPreviousHash() string {
	if x != nil {
		return x.PreviousHash
	}
	return ""
}

func (x *Block) GetHash() string {
	if x != nil {
		return x.Hash
	}
	return ""
}

func (x *Block) GetMerkleRoot() string {
	if x != nil {
		return x.MerkleRoot
	}
	return ""
}

func (x *Block) GetSignatures() []*BlockSignature {
	if x != nil {
		return x.Signatures
	}
	return nil
}

func (x *Transaction) GetTransactionId() string {
	if x != nil {
		return x.TransactionId
	}
	return ""
}

func (x *Transaction) GetAmount() float64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *Transaction) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

// Simplified gRPC service interface
type UnimplementedOrdererServiceServer struct{}

func (UnimplementedOrdererServiceServer) SubmitTx(ctx context.Context, req *SubmitTxRequest) (*SubmitTxReply, error) {
	return nil, fmt.Errorf("method SubmitTx not implemented")
}

func (UnimplementedOrdererServiceServer) StreamBlocks(req *StreamBlocksRequest, stream OrdererService_StreamBlocksServer) error {
	return fmt.Errorf("method StreamBlocks not implemented")
}

func (UnimplementedOrdererServiceServer) Consensus(ctx context.Context, msg *ConsensusMessage) (*AckReply, error) {
	return nil, fmt.Errorf("method Consensus not implemented")
}

type OrdererServiceServer interface {
	SubmitTx(ctx context.Context, req *SubmitTxRequest) (*SubmitTxReply, error)
	StreamBlocks(req *StreamBlocksRequest, stream OrdererService_StreamBlocksServer) error
	Consensus(ctx context.Context, msg *ConsensusMessage) (*AckReply, error)
}

type OrdererService_StreamBlocksServer interface {
	Send(*Block) error
	Context() context.Context
}

// gRPC Client interface
type OrdererServiceClient interface {
	SubmitTx(ctx context.Context, in *SubmitTxRequest, opts ...grpc.CallOption) (*SubmitTxReply, error)
	StreamBlocks(ctx context.Context, in *StreamBlocksRequest, opts ...grpc.CallOption) (OrdererService_StreamBlocksClient, error)
	Consensus(ctx context.Context, in *ConsensusMessage, opts ...grpc.CallOption) (*AckReply, error)
}

type OrdererService_StreamBlocksClient interface {
	Recv() (*Block, error)
	grpc.ClientStream
}

// NewOrdererServiceClient creates a new gRPC client
func NewOrdererServiceClient(cc grpc.ClientConnInterface) OrdererServiceClient {
	return &ordererServiceClient{cc}
}

type ordererServiceClient struct {
	cc grpc.ClientConnInterface
}

func (c *ordererServiceClient) SubmitTx(ctx context.Context, in *SubmitTxRequest, opts ...grpc.CallOption) (*SubmitTxReply, error) {
	out := new(SubmitTxReply)
	err := c.cc.Invoke(ctx, "/blockchain.OrdererService/SubmitTx", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ordererServiceClient) StreamBlocks(ctx context.Context, in *StreamBlocksRequest, opts ...grpc.CallOption) (OrdererService_StreamBlocksClient, error) {
	stream, err := c.cc.NewStream(ctx, &grpc.StreamDesc{
		StreamName:    "StreamBlocks",
		ServerStreams: true,
		ClientStreams: false,
	}, "/blockchain.OrdererService/StreamBlocks", opts...)
	if err != nil {
		return nil, err
	}
	x := &ordererServiceStreamBlocksClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ordererServiceStreamBlocksClient struct {
	grpc.ClientStream
}

func (x *ordererServiceStreamBlocksClient) Recv() (*Block, error) {
	m := new(Block)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *ordererServiceClient) Consensus(ctx context.Context, in *ConsensusMessage, opts ...grpc.CallOption) (*AckReply, error) {
	out := new(AckReply)
	err := c.cc.Invoke(ctx, "/blockchain.OrdererService/Consensus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Minimal interface implementations
func (x *ConsensusMessage) GetPrePrepare() *PrePrepareMessage {
	if x != nil {
		return x.PrePrepare
	}
	return nil
}

func (x *ConsensusMessage) GetPrepare() *PrepareMessage {
	if x != nil {
		return x.Prepare
	}
	return nil
}

func (x *ConsensusMessage) GetCommit() *CommitMessage {
	if x != nil {
		return x.Commit
	}
	return nil
}

func (x *ConsensusMessage) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

// gRPC service registration
func RegisterOrdererServiceServer(s grpc.ServiceRegistrar, srv OrdererServiceServer) {
	s.RegisterService(&OrdererService_ServiceDesc, srv)
}

var OrdererService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "blockchain.OrdererService",
	HandlerType: (*OrdererServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SubmitTx",
			Handler:    _OrdererService_SubmitTx_Handler,
		},
		{
			MethodName: "Consensus",
			Handler:    _OrdererService_Consensus_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamBlocks",
			Handler:       _OrdererService_StreamBlocks_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "blockchain.proto",
}

func _OrdererService_SubmitTx_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitTxRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdererServiceServer).SubmitTx(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/blockchain.OrdererService/SubmitTx",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdererServiceServer).SubmitTx(ctx, req.(*SubmitTxRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrdererService_StreamBlocks_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamBlocksRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(OrdererServiceServer).StreamBlocks(m, &ordererServiceStreamBlocksServer{ServerStream: stream})
}

type ordererServiceStreamBlocksServer struct {
	grpc.ServerStream
}

func (x *ordererServiceStreamBlocksServer) Send(m *Block) error {
	return x.ServerStream.SendMsg(m)
}

func _OrdererService_Consensus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConsensusMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdererServiceServer).Consensus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/blockchain.OrdererService/Consensus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdererServiceServer).Consensus(ctx, req.(*ConsensusMessage))
	}
	return interceptor(ctx, in, info, handler)
}
