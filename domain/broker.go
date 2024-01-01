package domain

import (
	"fmt"
	"geo-distributed-message-broker/data"
	"log/slog"
	"time"
)

type Broker interface {
	Publish(msg data.Message) (uint64, error)
	Subscribe(consumerName string, topic string) (<-chan data.Message, error)
	Unsubscribe(consumerName string, topic string)
	Acknowledge(consumerName string, msg data.Message) error
}

func NewBroker(repo data.Repository) Broker {
	slog.Info("Creating new broker 📬")

	return &broker{
		topics:        map[string]map[string]chan data.Message{},
		messagesCount: map[string]uint64{},
		repo:          repo,
	}
}

type broker struct {
	topics        map[string]map[string]chan data.Message // map[topic_name]map[consumer_name]chan data.Message
	messagesCount map[string]uint64                       // map[topic_name]uint64
	repo          data.Repository
}

func (b *broker) createTopic(topic string) error {
	slog.Info("Creating new topic", "topic", topic)
	count, err := b.repo.GetMessagesCount(topic)
	if err != nil {
		return err
	}

	b.topics[topic] = map[string]chan data.Message{}
	b.messagesCount[topic] = count

	return nil
}

func (b *broker) Publish(msg data.Message) (uint64, error) {
	slog.Debug("Publishing message", "topic", msg.Topic)

	// Create topic if it does not exist
	if _, ok := b.topics[msg.Topic]; !ok {
		if err := b.createTopic(msg.Topic); err != nil {
			return 0, err
		}
	}

	// Update messages count
	b.messagesCount[msg.Topic]++
	msg.ID = b.messagesCount[msg.Topic]

	// Save message to database
	err := b.repo.CreateMessage(&msg)
	if err != nil {
		return 0, err
	}

	// Send message to all consumers
	for _, consumer := range b.topics[msg.Topic] {
		consumer <- msg
	}

	return msg.ID, nil
}

func (b *broker) Subscribe(consumerName string, topic string) (<-chan data.Message, error) {
	slog.Debug("Subscribing to topic", "consumer", consumerName, "topic", topic)

	// Create topic if it does not exist
	if _, ok := b.topics[topic]; !ok {
		if err := b.createTopic(topic); err != nil {
			return nil, err
		}
	}

	// If consumer already subscribed to topic, return error
	if _, ok := b.topics[topic][consumerName]; ok {
		return nil, fmt.Errorf("consumer with name %s already exists for topic %s", consumerName, topic)
	}

	// Get all messages published after last consumed message
	messages, err := b.repo.GetMessages(consumerName, topic)
	if err != nil {
		return nil, err
	}

	// Create consumer
	consumer := make(chan data.Message, 10)
	b.topics[topic][consumerName] = consumer
	go func() {
		for _, msg := range messages {
			consumer <- msg
		}
	}()

	return consumer, nil
}

func (b *broker) Unsubscribe(consumerName string, topic string) {
	slog.Debug("Unsubscribing from topic", "consumer", consumerName, "topic", topic)

	if _, ok := b.topics[topic]; ok {
		delete(b.topics[topic], consumerName)
	}
}

func (b *broker) Acknowledge(consumerName string, msg data.Message) error {
	slog.Debug("Acknowledging message", "consumer", consumerName, "message", msg.ID)

	record := data.MessageConsumedRecord{
		MessageID:  msg.ID,
		Topic:      msg.Topic,
		ConsumedBy: consumerName,
		ConsumedAt: time.Now(),
	}

	return b.repo.CreateConsumedRecord(&record)
}
