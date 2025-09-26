package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Shopify/sarama"
	pb "shared/proto"
)

type KafkaProducer struct {
	producer sarama.SyncProducer
	scfTopic string
	auditTopic string
}

type TransactionMessage struct {
	TransactionId   string                 `json:"transaction_id"`
	TransactionType string                 `json:"transaction_type"`
	PeerId          string                 `json:"peer_id"`
	Channel         string                 `json:"channel"`
	Timestamp       int64                  `json:"timestamp"`
	Payload         map[string]interface{} `json:"payload"`
}

func NewKafkaProducer(brokers []string, scfTopic, auditTopic string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %v", err)
	}

	return &KafkaProducer{
		producer:  producer,
		scfTopic:  scfTopic,
		auditTopic: auditTopic,
	}, nil
}

func (kp *KafkaProducer) Close() error {
	return kp.producer.Close()
}

func (kp *KafkaProducer) PublishTransaction(ctx context.Context, transaction *pb.Transaction, channel string) error {
	// Determine topic based on channel
	topic := kp.scfTopic
	if channel == "audit" {
		topic = kp.auditTopic
	}

	// Convert protobuf transaction to JSON message
	message := TransactionMessage{
		TransactionId:   transaction.TransactionId,
		TransactionType: transaction.TransactionType,
		PeerId:          transaction.PeerId,
		Channel:         channel,
		Timestamp:       transaction.Timestamp.GetSeconds(),
		Payload: map[string]interface{}{
			"contract_id": transaction.ContractId,
			"token_id":    transaction.TokenId,
			"sender_id":   transaction.SenderId,
			"receiver_id": transaction.ReceiverId,
			"amount":      transaction.Amount,
			"payload":     transaction.Payload,
		},
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %v", err)
	}

	// Create Kafka message
	kafkaMsg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(transaction.TransactionId),
		Value: sarama.StringEncoder(jsonData),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("peer_id"),
				Value: []byte(transaction.PeerId),
			},
			{
				Key:   []byte("channel"),
				Value: []byte(channel),
			},
		},
	}

	// Send message
	partition, offset, err := kp.producer.SendMessage(kafkaMsg)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %v", err)
	}

	log.Printf("Published transaction %s to topic %s (partition: %d, offset: %d)",
		transaction.TransactionId, topic, partition, offset)

	return nil
}

func (kp *KafkaProducer) PublishEvent(ctx context.Context, event map[string]interface{}, channel string) error {
	// Determine topic based on channel
	topic := kp.scfTopic
	if channel == "audit" {
		topic = kp.auditTopic
	}

	// Add metadata to event
	event["channel"] = channel
	event["published_at"] = time.Now().Unix()

	// Serialize to JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	// Create Kafka message
	eventId, _ := event["eventId"].(string)
	kafkaMsg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(eventId),
		Value: sarama.StringEncoder(jsonData),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("event_type"),
				Value: []byte("blockchain_event"),
			},
			{
				Key:   []byte("channel"),
				Value: []byte(channel),
			},
		},
	}

	// Send message
	partition, offset, err := kp.producer.SendMessage(kafkaMsg)
	if err != nil {
		return fmt.Errorf("failed to send event to Kafka: %v", err)
	}

	log.Printf("Published event %s to topic %s (partition: %d, offset: %d)",
		eventId, topic, partition, offset)

	return nil
}

// Initialize Kafka producer for the peer node
func (p *PeerNode) initKafkaProducer() error {
	kafkaBrokersStr := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokersStr == "" {
		kafkaBrokersStr = "localhost:9092"
	}

	scfTopic := os.Getenv("SCF_CHANNEL_TOPIC")
	if scfTopic == "" {
		scfTopic = "scf-channel-tx"
	}

	auditTopic := os.Getenv("AUDIT_CHANNEL_TOPIC")
	if auditTopic == "" {
		auditTopic = "audit-channel-tx"
	}

	brokers := []string{kafkaBrokersStr}
	producer, err := NewKafkaProducer(brokers, scfTopic, auditTopic)
	if err != nil {
		return fmt.Errorf("failed to initialize Kafka producer: %v", err)
	}

	p.kafkaProducer = producer
	log.Printf("Initialized Kafka producer for peer %s with brokers: %v", p.nodeID, brokers)
	return nil
}

// Publish transaction to Kafka instead of calling Orderer directly
func (p *PeerNode) publishTransactionToKafka(transaction *pb.Transaction, channel string) error {
	if p.kafkaProducer == nil {
		return fmt.Errorf("Kafka producer not initialized")
	}

	ctx := context.Background()
	return p.kafkaProducer.PublishTransaction(ctx, transaction, channel)
}

// Publish blockchain event to Kafka
func (p *PeerNode) publishEventToKafka(event map[string]interface{}, channel string) error {
	if p.kafkaProducer == nil {
		return fmt.Errorf("Kafka producer not initialized")
	}

	ctx := context.Background()
	return p.kafkaProducer.PublishEvent(ctx, event, channel)
}
