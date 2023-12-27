package domain

import (
	"fmt"
	"geo-distributed-message-broker/data"
	"log/slog"
	"time"
)

type Broker interface {
	Publish(msg data.Message) error
	Subscribe(name string, topic string) (<-chan data.Message, error)
	Unsubscribe(name string, topic string)
	Acknowledge(name string, msg data.Message) error
}

func NewBroker(repo data.Repository) Broker {
	slog.Info("Creating new broker")

	return &broker{
		topics: map[string]map[string]chan data.Message{},
		repo:   repo,
	}
}

type broker struct {
	topics map[string]map[string]chan data.Message // map[topic_name]map[consumer_name]chan data.Message
	repo   data.Repository
}

func (b *broker) Publish(msg data.Message) error {
	// Save message to database
	err := b.repo.CreateMessage(&msg)
	if err != nil {
		return err
	}

	// Send message to all consumers
	if _, ok := b.topics[msg.Topic]; ok {
		for _, consumer := range b.topics[msg.Topic] {
			consumer <- msg
		}
	}

	return nil
}

func (b *broker) Subscribe(name string, topic string) (<-chan data.Message, error) {
	// Create topic if it does not exist
	if _, ok := b.topics[topic]; !ok {
		b.topics[topic] = map[string]chan data.Message{}
	}

	// If consumer already subscribed to topic, return error
	if _, ok := b.topics[topic][name]; ok {
		return nil, fmt.Errorf("consumer with name %s already exists for topic %s", name, topic)
	}

	// Get all messages published after last consumed message
	messages, err := b.repo.GetMessages(name, topic)
	if err != nil {
		return nil, err
	}

	// Create consumer
	consumer := make(chan data.Message, 10)
	b.topics[topic][name] = consumer
	go func() {
		for _, msg := range messages {
			consumer <- msg
		}
	}()

	return consumer, nil
}

func (b *broker) Unsubscribe(name string, topic string) {
	if _, ok := b.topics[topic]; ok {
		delete(b.topics[topic], name)
	}
}

func (b *broker) Acknowledge(name string, msg data.Message) error {
	record := data.MessageConsumedRecord{
		MessageID:  msg.ID,
		Topic:      msg.Topic,
		ConsumedBy: name,
		ConsumedAt: time.Now(),
	}

	return b.repo.CreateConsumedRecord(&record)
}
