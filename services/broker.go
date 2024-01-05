package services

import (
	"fmt"
	"geo-distributed-message-broker/data"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

type BrokerService interface {
	Publish(msg data.Message) (string, error)
	Subscribe(consumerName string, topics []string) (<-chan data.Message, error)
	Unsubscribe(consumerName string, topics []string)
	Acknowledge(consumerName string, msg data.Message) error
}

func NewBrokerService(repo data.Repository) BrokerService {
	slog.Info("Creating new broker 📬")

	return &brokerService{
		topics: map[string]map[string]chan data.Message{},
		repo:   repo,
	}
}

type brokerService struct {
	topics map[string]map[string]chan data.Message // map[topic_name]map[consumer_name]chan data.Message
	mu     sync.RWMutex                            // protects topics
	repo   data.Repository
}

func (b *brokerService) Publish(msg data.Message) (string, error) {
	// Generate message ID if not provided
	if msg.ID == "" {
		msg.ID = uuid.NewString()
	}

	// Set timestamp if not provided
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UnixMicro()
	}

	slog.Debug("Publishing message", "topic", msg.Topic, "message", msg.ID)

	// Create topic if it does not exist
	b.mu.Lock()
	if _, ok := b.topics[msg.Topic]; !ok {
		b.topics[msg.Topic] = map[string]chan data.Message{}
	}
	b.mu.Unlock()

	err := b.repo.CreateMessage(&msg)
	if err != nil {
		return "", err
	}

	// Send message to all consumers
	b.mu.RLock()
	for _, consumer := range b.topics[msg.Topic] {
		consumer <- msg
	}
	b.mu.RUnlock()

	return msg.ID, nil
}

func (b *brokerService) Subscribe(consumerName string, topics []string) (<-chan data.Message, error) {
	slog.Debug("Subscribing to topic", "consumer", consumerName, "topics", topics)

	for _, topic := range topics {
		// Create topic if it does not exist
		b.mu.Lock()
		if _, ok := b.topics[topic]; !ok {
			b.topics[topic] = map[string]chan data.Message{}
		}
		b.mu.Unlock()

		// If consumer already subscribed to topic, return error
		if _, ok := b.topics[topic][consumerName]; ok {
			return nil, fmt.Errorf("consumer with name %s already exists for topics %v", consumerName, topics)
		}
	}

	// Create consumer
	consumer := make(chan data.Message, 10)

	for _, topic := range topics {
		// Get all messages published after last consumed message
		messages, err := b.repo.GetMessages(consumerName, topic)
		if err != nil {
			return nil, err
		}

		// Add consumer to topic
		b.mu.Lock()
		b.topics[topic][consumerName] = consumer
		b.mu.Unlock()

		go func() {
			for _, msg := range messages {
				consumer <- msg
			}
		}()
	}

	return consumer, nil
}

func (b *brokerService) Unsubscribe(consumerName string, topics []string) {
	slog.Debug("Unsubscribing from topic", "consumer", consumerName, "topics", topics)

	b.mu.Lock()
	for _, topic := range topics {
		if _, ok := b.topics[topic]; ok {
			delete(b.topics[topic], consumerName)
		}
	}
	b.mu.Unlock()
}

func (b *brokerService) Acknowledge(consumerName string, msg data.Message) error {
	slog.Debug("Acknowledging message", "consumer", consumerName, "topic", msg.Topic, "message", msg.ID)

	record := data.MessageConsumedRecord{
		MessageID:  msg.ID,
		ConsumedBy: consumerName,
		Timestamp:  msg.Timestamp,
		Topic:      msg.Topic,
	}

	return b.repo.CreateConsumedRecord(&record)
}
