package services

import (
	"geo-distributed-message-broker/data"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

type BrokerService interface {
	Publish(msg data.Message) (string, error)
	Subscribe(topics map[string]int64) (<-chan data.Message, string, error)
	Unsubscribe(subscriberID string, topics map[string]int64)
}

func NewBrokerService(repo data.Repository) BrokerService {
	slog.Info("Creating new broker 📬")

	return &brokerService{
		topics: map[string]map[string]chan data.Message{},
		repo:   repo,
	}
}

type brokerService struct {
	topics map[string]map[string]chan data.Message // map[topic_name]map[subscriber_id]chan data.Message
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

	// Send message to all subscribers
	b.mu.RLock()
	for _, subscriber := range b.topics[msg.Topic] {
		subscriber <- msg
	}
	b.mu.RUnlock()

	return msg.ID, nil
}

func (b *brokerService) Subscribe(topics map[string]int64) (<-chan data.Message, string, error) {
	// Generate subscriber ID
	subscriberID := uuid.NewString()

	slog.Debug("Subscribing to topics", "subscriber", subscriberID, "topics", topics)

	b.mu.Lock()
	for topic := range topics {
		// Create topic if it does not exist
		if _, ok := b.topics[topic]; !ok {
			b.topics[topic] = map[string]chan data.Message{}
		}
	}
	b.mu.Unlock()

	// Create subscriber channel
	subscriber := make(chan data.Message, 10)

	for topic, timestamp := range topics {
		// Get all messages published after last timestamp
		msgChan := b.repo.GetMessages(topic, timestamp)

		// Add subscriber to topic
		b.mu.Lock()
		b.topics[topic][subscriberID] = subscriber
		b.mu.Unlock()

		go func() {
			for {
				messages, ok := <-msgChan

				if !ok {
					break
				}

				for _, msg := range messages {
					subscriber <- msg
				}
			}
		}()
	}

	return subscriber, subscriberID, nil
}

func (b *brokerService) Unsubscribe(subscriberID string, topics map[string]int64) {
	slog.Debug("Unsubscribing from topics", "subscriber", subscriberID, "topics", topics)

	b.mu.Lock()
	for topic := range topics {
		if _, ok := b.topics[topic]; ok {
			delete(b.topics[topic], subscriberID)
		}
	}
	b.mu.Unlock()
}
