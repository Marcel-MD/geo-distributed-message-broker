package data

import (
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	CreateMessage(message *Message) error
	GetMessages(topicName string, timestamp int64) <-chan []Message
}

func NewRepository(db *gorm.DB) Repository {
	slog.Info("Creating new repository 🗄️")

	return &repository{db: db}
}

type repository struct {
	db *gorm.DB
}

func (r *repository) CreateMessage(message *Message) error {
	return r.db.Create(message).Error
}

func (r *repository) GetMessages(topicName string, timestamp int64) <-chan []Message {
	msgChan := make(chan []Message, 1)

	go func() {
		var messages []Message

		// Processing records in batches of 50
		result := r.db.Where("topic = ? AND timestamp > ?", topicName, timestamp).Order("timestamp").FindInBatches(&messages, 50, func(tx *gorm.DB, batch int) error {
			select {
			case msgChan <- messages:
				slog.Debug("Sending batch of messages", "topic", topicName, "batch", batch, "size", len(messages))

			case <-time.After(1 * time.Second):
				slog.Error("Timeout while getting messages from database")
				return errors.New("timeout while getting messages from database")
			}

			return nil
		})

		close(msgChan)

		if result.Error != nil {
			slog.Error("Error while getting messages from database", "error", result.Error)
		}
	}()

	return msgChan
}
