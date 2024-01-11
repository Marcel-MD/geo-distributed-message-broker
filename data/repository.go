package data

import (
	"log/slog"

	"gorm.io/gorm"
)

type Repository interface {
	CreateMessage(message *Message) error
	GetMessages(topicName string, timestamp int64) ([]Message, error)
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

func (r *repository) GetMessages(topicName string, timestamp int64) ([]Message, error) {
	var messages []Message
	if err := r.db.Where("topic = ? AND timestamp > ?", topicName, timestamp).Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}
