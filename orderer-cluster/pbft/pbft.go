package pbft

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"
	"sync"
	"time"

	"orderer-cluster/proto"

	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// PBFT consensus states
const (
	PRE_PREPARE = "PRE_PREPARE"
	PREPARE     = "PREPARE"
	COMMIT      = "COMMIT"
)

// PBFTNode represents a PBFT consensus node
type PBFTNode struct {
	nodeID       string
	view         int
	sequence     int64
	f            int // max faulty nodes (n = 3f + 1)
	privateKey   *ecdsa.PrivateKey
	publicKey    *ecdsa.PublicKey
	orderers     map[string]*OrdererInfo // orderer_id -> info
	mempool      Mempool
	blockBuilder BlockBuilder
	db           *mongo.Database

	// Consensus state
	currentBlock   *proto.Block
	prePrepareMsgs map[string]*proto.PrePrepareMessage         // digest -> message
	prepareMsgs    map[string]map[string]*proto.PrepareMessage // digest -> orderer_id -> message
	commitMsgs     map[string]map[string]*proto.CommitMessage  // digest -> orderer_id -> message

	// Communication channels
	consensusCh chan *proto.ConsensusMessage
	blockCh     chan *proto.Block

	// gRPC clients for inter-orderer communication
	ordererClients map[string]proto.OrdererServiceClient
	ordererConns   map[string]*grpc.ClientConn

	mutex sync.RWMutex
}

// BlockBuilder interface - will be implemented by block.BlockBuilder
type BlockBuilder interface {
	CreateBlock(transactions []*proto.Transaction, height int64) *proto.Block
	GetLatestBlockHeight() int64
	GetBlockByHeight(height int64) (*proto.Block, error)
}

// Mempool interface - will be implemented by mempool.Mempool
type Mempool interface {
	AddTransaction(tx *proto.Transaction) error
	GetPendingTransactions() []*proto.Transaction
	RemoveTransactions(txs []*proto.Transaction)
	HasPendingTransactions() bool
}

// OrdererInfo holds information about other orderers
type OrdererInfo struct {
	ID        string
	PublicKey *ecdsa.PublicKey
	Address   string
}

// NewPBFTNode creates a new PBFT consensus node
func NewPBFTNode(nodeID string, f int, privateKey *ecdsa.PrivateKey, orderers map[string]*OrdererInfo,
	mempool Mempool, blockBuilder BlockBuilder, db *mongo.Database) *PBFTNode {

	node := &PBFTNode{
		nodeID:         nodeID,
		view:           0,
		sequence:       0,
		f:              f,
		privateKey:     privateKey,
		publicKey:      &privateKey.PublicKey,
		orderers:       orderers,
		mempool:        mempool,
		blockBuilder:   blockBuilder,
		db:             db,
		prePrepareMsgs: make(map[string]*proto.PrePrepareMessage),
		prepareMsgs:    make(map[string]map[string]*proto.PrepareMessage),
		commitMsgs:     make(map[string]map[string]*proto.CommitMessage),
		consensusCh:    make(chan *proto.ConsensusMessage, 100),
		blockCh:        make(chan *proto.Block, 10),
		ordererClients: make(map[string]proto.OrdererServiceClient),
		ordererConns:   make(map[string]*grpc.ClientConn),
	}

	// Initialize gRPC clients to other orderers
	for id, orderer := range orderers {
		if id != nodeID { // Don't connect to self
			conn, err := grpc.Dial(orderer.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Printf("Failed to connect to orderer %s at %s: %v", id, orderer.Address, err)
				continue
			}
			client := proto.NewOrdererServiceClient(conn)
			node.ordererClients[id] = client
			node.ordererConns[id] = conn
			log.Printf("Connected to orderer %s at %s", id, orderer.Address)
		}
	}

	return node
}

// Start begins the PBFT consensus process
func (p *PBFTNode) Start(ctx context.Context) {
	log.Printf("PBFT Node %s starting consensus engine", p.nodeID)

	// Start consensus message handler
	go p.handleConsensusMessages(ctx)

	// Start block creation timer
	go p.blockCreationTimer(ctx)
}

// SubmitTransaction adds a transaction to the mempool and finalizes block immediately if primary
func (p *PBFTNode) SubmitTransaction(tx *proto.Transaction) error {
	// Add to mempool
	if err := p.mempool.AddTransaction(tx); err != nil {
		return err
	}

	// If this node is primary, finalize block immediately (simplified consensus for demo)
	if p.isPrimary() {
		go p.finalizeBlockImmediately()
	}

	return nil
}

// startConsensus initiates PBFT consensus for pending transactions
func (p *PBFTNode) startConsensus() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Check if consensus is already in progress
	if p.currentBlock != nil {
		log.Printf("Consensus already in progress for block height %d", p.currentBlock.Height)
		return
	}

	// Get transactions from mempool
	txs := p.mempool.GetPendingTransactions()
	if len(txs) == 0 {
		return
	}

	// Create block
	block := p.blockBuilder.CreateBlock(txs, p.sequence+1)
	p.currentBlock = block

	// Calculate digest
	digest := p.calculateBlockDigest(block)

	// Send Pre-Prepare message
	prePrepareMsg := &proto.PrePrepareMessage{
		View:           fmt.Sprintf("%d", p.view),
		SequenceNumber: fmt.Sprintf("%d", p.sequence+1),
		Block:          block,
		Digest:         digest,
	}

	consensusMsg := &proto.ConsensusMessage{
		Type: PRE_PREPARE,
		Message: &proto.ConsensusMessage_PrePrepare{
			PrePrepare: prePrepareMsg,
		},
	}

	// Broadcast to all orderers
	p.broadcastConsensusMessage(consensusMsg)

	log.Printf("Primary %s initiated consensus for block height %d", p.nodeID, block.Height)
}

// handleConsensusMessages processes incoming consensus messages
func (p *PBFTNode) handleConsensusMessages(ctx context.Context) {
	for {
		select {
		case msg := <-p.consensusCh:
			p.processConsensusMessage(msg)
		case <-ctx.Done():
			return
		}
	}
}

// processConsensusMessage handles different types of consensus messages
func (p *PBFTNode) processConsensusMessage(msg *proto.ConsensusMessage) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	switch msg.Type {
	case PRE_PREPARE:
		p.handlePrePrepare(msg.GetPrePrepare())
	case PREPARE:
		p.handlePrepare(msg.GetPrepare())
	case COMMIT:
		p.handleCommit(msg.GetCommit())
	}
}

// handlePrePrepare processes Pre-Prepare messages
func (p *PBFTNode) handlePrePrepare(msg *proto.PrePrepareMessage) {
	digest := msg.Digest

	// Store Pre-Prepare message
	p.prePrepareMsgs[digest] = msg

	// Verify block digest
	if !p.verifyBlockDigest(msg.Block, digest) {
		log.Printf("Invalid block digest in Pre-Prepare from primary")
		return
	}

	// Initialize prepare messages map for this digest
	if p.prepareMsgs[digest] == nil {
		p.prepareMsgs[digest] = make(map[string]*proto.PrepareMessage)
	}

	// Send Prepare message to ourselves (simulate receiving our own Prepare)
	prepareMsg := &proto.PrepareMessage{
		View:           msg.View,
		SequenceNumber: msg.SequenceNumber,
		Digest:         digest,
		OrdererId:      p.nodeID,
	}
	p.prepareMsgs[digest][p.nodeID] = prepareMsg

	// Broadcast Prepare message to other orderers
	consensusMsg := &proto.ConsensusMessage{
		Type: PREPARE,
		Message: &proto.ConsensusMessage_Prepare{
			Prepare: prepareMsg,
		},
	}

	p.broadcastConsensusMessage(consensusMsg)

	log.Printf("Node %s sent Prepare for digest %s", p.nodeID, digest)
}

// handlePrepare processes Prepare messages
func (p *PBFTNode) handlePrepare(msg *proto.PrepareMessage) {
	digest := msg.Digest

	// Initialize prepare messages map for this digest if needed
	if p.prepareMsgs[digest] == nil {
		p.prepareMsgs[digest] = make(map[string]*proto.PrepareMessage)
	}

	// Store Prepare message
	p.prepareMsgs[digest][msg.OrdererId] = msg

	// If this is our own Prepare message, don't send Commit yet
	if msg.OrdererId == p.nodeID {
		log.Printf("Node %s received own Prepare message for digest %s", p.nodeID, digest)
		return
	}

	// Check if we have 2f + 1 Prepare messages (including our own)
	// In PBFT, we need 2f+1 total Prepare messages including our own
	log.Printf("Node %s has %d Prepare messages for digest %s", p.nodeID, len(p.prepareMsgs[digest]), digest)
	if len(p.prepareMsgs[digest]) >= 2*p.f+1 {
		log.Printf("Node %s has enough Prepare messages, sending Commit for digest %s", p.nodeID, digest)
		// Initialize commit messages map for this digest
		if p.commitMsgs[digest] == nil {
			p.commitMsgs[digest] = make(map[string]*proto.CommitMessage)
		}

		// Send Commit message to ourselves (simulate receiving our own Commit)
		commitMsg := &proto.CommitMessage{
			View:           msg.View,
			SequenceNumber: msg.SequenceNumber,
			Digest:         digest,
			OrdererId:      p.nodeID,
		}
		p.commitMsgs[digest][p.nodeID] = commitMsg

		// Broadcast Commit message to other orderers
		consensusMsg := &proto.ConsensusMessage{
			Type: COMMIT,
			Message: &proto.ConsensusMessage_Commit{
				Commit: commitMsg,
			},
		}

		p.broadcastConsensusMessage(consensusMsg)

		log.Printf("Node %s sent Commit for digest %s", p.nodeID, digest)
	}
}

// handleCommit processes Commit messages
func (p *PBFTNode) handleCommit(msg *proto.CommitMessage) {
	digest := msg.Digest

	// Initialize commit messages map for this digest if needed
	if p.commitMsgs[digest] == nil {
		p.commitMsgs[digest] = make(map[string]*proto.CommitMessage)
	}

	// Store Commit message
	p.commitMsgs[digest][msg.OrdererId] = msg

	// If this is our own Commit message, don't finalize yet
	if msg.OrdererId == p.nodeID {
		log.Printf("Node %s received own Commit message for digest %s", p.nodeID, digest)
		return
	}

	// Check if we have 2f + 1 Commit messages (including our own)
	if len(p.commitMsgs[digest]) >= 2*p.f+1 {
		// Consensus reached! Finalize block
		p.finalizeBlock(digest)
	}
}

// finalizeBlockImmediately creates and finalizes a block immediately (simplified consensus)
func (p *PBFTNode) finalizeBlockImmediately() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Get transactions from mempool
	txs := p.mempool.GetPendingTransactions()
	if len(txs) == 0 {
		return
	}

	// Create block
	block := p.blockBuilder.CreateBlock(txs, p.sequence+1)

	// Add dummy signatures from all orderers (simplified for demo)
	var signatures []*proto.BlockSignature
	for ordererID := range p.orderers {
		signature := p.signBlock(block, ordererID)
		signatures = append(signatures, signature)
	}
	block.Signatures = signatures

	// Persist block to database
	if err := p.persistBlock(block); err != nil {
		log.Printf("Failed to persist block: %v", err)
		return
	}

	// Broadcast finalized block
	p.blockCh <- block

	// Update sequence number
	p.sequence = block.Height

	// Remove transactions from mempool
	p.mempool.RemoveTransactions(txs)

	log.Printf("Block %d finalized immediately with %d signatures", block.Height, len(signatures))
}

// finalizeBlock commits the block to the blockchain
func (p *PBFTNode) finalizeBlock(digest string) {
	prePrepareMsg := p.prePrepareMsgs[digest]
	if prePrepareMsg == nil {
		log.Printf("No Pre-Prepare message found for digest %s", digest)
		return
	}

	block := prePrepareMsg.Block

	// Add signatures from all orderers who sent Commit messages
	var signatures []*proto.BlockSignature
	for ordererID := range p.commitMsgs[digest] {
		if _, exists := p.orderers[ordererID]; exists {
			signature := p.signBlock(block, ordererID)
			signatures = append(signatures, signature)
		}
	}

	block.Signatures = signatures

	// Persist block to database
	if err := p.persistBlock(block); err != nil {
		log.Printf("Failed to persist block: %v", err)
		return
	}

	// Broadcast finalized block
	p.blockCh <- block

	// Clean up consensus state
	delete(p.prePrepareMsgs, digest)
	delete(p.prepareMsgs, digest)
	delete(p.commitMsgs, digest)

	// Update sequence number
	p.sequence = block.Height

	// Remove transactions from mempool
	p.mempool.RemoveTransactions(block.Transactions)

	log.Printf("Block %d finalized with %d signatures", block.Height, len(signatures))
}

// broadcastConsensusMessage sends consensus message to all orderers
func (p *PBFTNode) broadcastConsensusMessage(msg *proto.ConsensusMessage) {
	log.Printf("Broadcasting %s message from %s", msg.Type, p.nodeID)

	// Send to all other orderers via gRPC
	for ordererID, client := range p.ordererClients {
		go func(id string, c proto.OrdererServiceClient) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := c.Consensus(ctx, msg)
			if err != nil {
				log.Printf("Failed to send %s message to orderer %s: %v", msg.Type, id, err)
			} else {
				log.Printf("Sent %s message to orderer %s", msg.Type, id)
			}
		}(ordererID, client)
	}
}

// ReceiveConsensusMessage handles incoming consensus messages from other orderers
func (p *PBFTNode) ReceiveConsensusMessage(msg *proto.ConsensusMessage) {
	select {
	case p.consensusCh <- msg:
	default:
		log.Printf("Consensus message channel full, dropping message")
	}
}

// GetBlocksChannel returns channel for receiving finalized blocks
func (p *PBFTNode) GetBlocksChannel() <-chan *proto.Block {
	return p.blockCh
}

// Close closes all gRPC connections
func (p *PBFTNode) Close() {
	for _, conn := range p.ordererConns {
		conn.Close()
	}
}

// Utility methods

func (p *PBFTNode) isPrimary() bool {
	return p.nodeID == "ord1" // ord1 is always primary in view 0
}

func (p *PBFTNode) calculateBlockDigest(block *proto.Block) string {
	// Simple digest calculation - in production use proper crypto
	data := fmt.Sprintf("%d-%s-%s", block.Height, block.PreviousHash, block.MerkleRoot)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (p *PBFTNode) verifyBlockDigest(block *proto.Block, digest string) bool {
	expectedDigest := p.calculateBlockDigest(block)
	return expectedDigest == digest
}

func (p *PBFTNode) signBlock(block *proto.Block, ordererID string) *proto.BlockSignature {
	// Sign block hash with orderer's private key
	blockData := fmt.Sprintf("%d-%s-%s", block.Height, block.Hash, block.MerkleRoot)
	hash := sha256.Sum256([]byte(blockData))

	r, s, err := ecdsa.Sign(rand.Reader, p.privateKey, hash[:])
	if err != nil {
		log.Printf("Failed to sign block: %v", err)
		return nil
	}

	signature := r.Bytes()
	signature = append(signature, s.Bytes()...)

	// Encode public key
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(p.publicKey)
	if err != nil {
		log.Printf("Failed to marshal public key: %v", err)
		return nil
	}

	pubKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return &proto.BlockSignature{
		OrdererId: ordererID,
		Signature: signature,
		PublicKey: string(pubKeyPem),
	}
}

func (p *PBFTNode) persistBlock(block *proto.Block) error {
	// Persist to MongoDB shared database
	collection := p.db.Collection("blocks")

	_, err := collection.InsertOne(context.Background(), block)
	return err
}

func (p *PBFTNode) blockCreationTimer(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if p.isPrimary() && p.mempool.HasPendingTransactions() {
				p.startConsensus()
			}
		case <-ctx.Done():
			return
		}
	}
}
