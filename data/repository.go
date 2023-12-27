package data

import (
	"log/slog"

	"gorm.io/gorm"
)

type Repository interface {
	CreateMessage(message *Message) error
	CreateConsumedRecord(record *MessageConsumedRecord) error
	GetMessages(consumerName string, topicName string) ([]Message, error)
}

func NewRepository(db *gorm.DB) Repository {
	slog.Info("Creating new repository")

	return &repository{db: db}
}

type repository struct {
	db *gorm.DB
}

func (r *repository) CreateMessage(message *Message) error {
	return r.db.Create(message).Error
}

func (r *repository) CreateConsumedRecord(record *MessageConsumedRecord) error {
	return r.db.Create(record).Error
}

func (r *repository) GetMessages(consumerName string, topicName string) ([]Message, error) {
	var lastConsumedRecord MessageConsumedRecord
	err := r.db.Where("topic = ? AND consumed_by = ?", topicName, consumerName).Last(&lastConsumedRecord).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	var messages []Message
	if err := r.db.Where("topic = ? AND published_at > ?", topicName, lastConsumedRecord.ConsumedAt).Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}
