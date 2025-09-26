package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Simple block structure
type Block struct {
	BlockNumber  int64
	Timestamp    int64
	Transactions []Transaction
	PreviousHash string
	Hash         string
	PeerID       string
	OrdererID    string
	MerkleRoot   string
}

// Simple transaction structure
type Transaction struct {
	TransactionID   string
	TransactionType string
	ContractID      string
	TokenID         string
	SenderID        string
	ReceiverID      string
	Amount          float64
	Payload         string
	Timestamp       int64
}

func generateEventID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

type OrdererNode struct {
	nodeID         string
	port           int
	peers          []string
	kafkaBrokers   []string
	scfTopic       string
	auditTopic     string
	blockBuffer    chan *Block
	eventBuffer    chan map[string]interface{}
	mutex          sync.Mutex
	currentBlock   *Block
	kafkaConsumer  sarama.ConsumerGroup
	db             *mongo.Database
}

type OrdererService struct {
	orderer *OrdererNode
}

func NewOrdererNode(nodeID string, port int, peers []string, kafkaBrokers []string, scfTopic, auditTopic string) *OrdererNode {
	return &OrdererNode{
		nodeID:       nodeID,
		port:         port,
		peers:        peers,
		kafkaBrokers: kafkaBrokers,
		scfTopic:     scfTopic,
		auditTopic:   auditTopic,
		blockBuffer:  make(chan *Block, 100),
		eventBuffer:  make(chan map[string]interface{}, 1000),
	}
}

func (o *OrdererNode) Start() {
	log.Printf("Orderer %s starting initialization...", o.nodeID)

	// Connect to MongoDB
	if err := o.connectToDatabase(); err != nil {
		log.Printf("Warning: Failed to connect to MongoDB: %v", err)
	} else {
		log.Printf("Successfully connected to MongoDB")
	}

	// Initialize Kafka consumer
	if err := o.initKafkaConsumer(); err != nil {
		log.Printf("Warning: Failed to initialize Kafka consumer: %v", err)
	}

	// Start event processing goroutine
	go o.processEvents()

	// Start block ordering goroutine
	go o.orderBlocks()

	log.Printf("Orderer %s started on port %d", o.nodeID, o.port)

	// Keep the service running
	select {}
}

func (o *OrdererNode) orderBlocks() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case block := <-o.blockBuffer:
			o.processBlock(block)
		case <-ticker.C:
			o.createOrderedBlock()
		}
	}
}

func (o *OrdererNode) processBlock(block *Block) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	log.Printf("Orderer %s processing block %d", o.nodeID, block.BlockNumber)

	// Validate block and add to ordering queue
	// In a real implementation, this would involve consensus among orderer nodes

	// Broadcast to other orderer nodes for consensus
	o.broadcastToPeers(block)
}

func (o *OrdererNode) createOrderedBlock() {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.currentBlock == nil {
		return
	}

	// Create ordered block with consensus
	orderedBlock := &Block{
		BlockNumber:  o.currentBlock.BlockNumber,
		Timestamp:    time.Now().Unix(),
		Transactions: o.currentBlock.Transactions,
		PreviousHash: o.currentBlock.PreviousHash,
		Hash:         o.calculateBlockHash(o.currentBlock),
		OrdererID:    o.nodeID,
	}

	log.Printf("Orderer %s created ordered block %d", o.nodeID, orderedBlock.BlockNumber)

	// Broadcast ordered block to peers
	o.broadcastOrderedBlock(orderedBlock)

	o.currentBlock = nil
}

func (o *OrdererNode) connectToDatabase() error {
	// Get MongoDB connection string from environment
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		// Default for orderer - connect to shared database
		mongoURI = "mongodb://root:example@mongo-shared:27017/blockchain?authSource=admin"
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		return err
	}

	// Ping the database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	// Get database
	dbName := "blockchain" // shared blockchain database
	o.db = client.Database(dbName)

	log.Printf("Orderer connected to MongoDB database: %s", dbName)
	return nil
}

func (o *OrdererNode) initKafkaConsumer() error {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	groupID := fmt.Sprintf("orderer-%s-group", o.nodeID)
	consumerGroup, err := sarama.NewConsumerGroup(o.kafkaBrokers, groupID, config)
	if err != nil {
		return fmt.Errorf("failed to create consumer group: %v", err)
	}

	o.kafkaConsumer = consumerGroup

	// Start consuming in background
	go o.consumeTopics()

	log.Printf("Initialized Kafka consumer for orderer %s with brokers: %v", o.nodeID, o.kafkaBrokers)
	return nil
}

func (o *OrdererNode) consumeTopics() {
	topics := []string{o.scfTopic, o.auditTopic}

	handler := &ConsumerGroupHandler{
		orderer: o,
		ready:   make(chan bool),
	}

	ctx := context.Background()
	for {
		err := o.kafkaConsumer.Consume(ctx, topics, handler)
		if err != nil {
			log.Printf("Error from consumer: %v", err)
		}
		if ctx.Err() != nil {
			return
		}
		handler.ready = make(chan bool)
	}
}

func (o *OrdererNode) processEvents() {
	for event := range o.eventBuffer {
		log.Printf("Orderer %s processing event: %v", o.nodeID, event)

		// Convert event to transaction and add to processing queue
		// For now, we'll just log the event
		o.processEvent(event)
	}
}

func (o *OrdererNode) processEvent(event map[string]interface{}) {
	// Process the blockchain event from Kafka
	// This involves validation, ordering, and block creation
	log.Printf("Processing event type: %v", event["eventType"])

	// Store event in database for persistence
	eventDoc := map[string]interface{}{
		"eventId":    generateEventID(),
		"eventType":  event["eventType"],
		"data":       event["data"],
		"timestamp":  event["timestamp"],
		"processed":  true,
		"ordererId":  o.nodeID,
	}

	// Save to database
	if o.db != nil {
		_, err := o.db.Collection("events").InsertOne(context.Background(), eventDoc)
		if err != nil {
			log.Printf("Failed to save event to database: %v", err)
		} else {
			log.Printf("Event saved to database: %s", eventDoc["eventId"])
		}
	} else {
		log.Printf("Database not connected, would save event: %+v", eventDoc)
	}
}

type ConsumerGroupHandler struct {
	orderer *OrdererNode
	ready   chan bool
}

func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		h.processMessage(message)
		session.MarkMessage(message, "")
	}
	return nil
}

func (h *ConsumerGroupHandler) processMessage(message *sarama.ConsumerMessage) {
	var event map[string]interface{}
	if err := json.Unmarshal(message.Value, &event); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	log.Printf("Received event from topic %s: %v", message.Topic, event["eventType"])

	// Send to event processing channel
	select {
	case h.orderer.eventBuffer <- event:
	default:
		log.Printf("Event buffer full, dropping event")
	}
}

func (o *OrdererNode) broadcastToPeers(block *Block) {
	// Implementation for broadcasting to other orderer nodes
	log.Printf("Orderer %s broadcasting block to peers", o.nodeID)
}

func (o *OrdererNode) broadcastOrderedBlock(block *Block) {
	// Implementation for broadcasting ordered block to connected peers
	log.Printf("Orderer %s broadcasting ordered block %d", o.nodeID, block.BlockNumber)
}

func (o *OrdererNode) calculateBlockHash(block *Block) string {
	// Simple hash calculation - in real implementation use proper crypto
	return fmt.Sprintf("hash_%d_%s", block.BlockNumber, o.nodeID)
}

// Simplified service methods - no gRPC for now
func (s *OrdererService) SubmitBlock(peerID string, block *Block) (bool, string) {
	log.Printf("Orderer %s received block %d from peer %s", s.orderer.nodeID, block.BlockNumber, peerID)

	// Add to ordering queue
	s.orderer.blockBuffer <- block

	return true, "Block submitted for ordering"
}

func (s *OrdererService) GetOrderedBlocks(peerID string, lastBlockNumber int64) []*Block {
	// Return ordered blocks for the requesting peer
	return []*Block{} // Implement proper block retrieval
}

func main() {
	nodeID := os.Getenv("ORDERER_NODE_ID")
	if nodeID == "" {
		nodeID = "ord1"
	}

	portStr := os.Getenv("ORDERER_PORT")
	port := 7050 // default port
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	kafkaBrokersStr := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokersStr == "" {
		kafkaBrokersStr = "kafka:29092"
	}
	kafkaBrokers := []string{kafkaBrokersStr}

	scfTopic := os.Getenv("SCF_CHANNEL_TOPIC")
	if scfTopic == "" {
		scfTopic = "scf-channel-tx"
	}

	auditTopic := os.Getenv("AUDIT_CHANNEL_TOPIC")
	if auditTopic == "" {
		auditTopic = "audit-channel-tx"
	}

	// Configure peers based on node ID
	var peers []string
	switch nodeID {
	case "ord1":
		peers = []string{"ord2:7060", "ord3:7070"}
	case "ord2":
		peers = []string{"ord1:7050", "ord3:7070"}
	case "ord3":
		peers = []string{"ord1:7050", "ord2:7060"}
	}

	orderer := NewOrdererNode(nodeID, port, peers, kafkaBrokers, scfTopic, auditTopic)
	orderer.Start()
}
