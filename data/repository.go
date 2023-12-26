package data

import (
	"gorm.io/gorm"
)

type Repository interface {
	CreateMessage(message *Message) error
	CreateConsumedRecord(record *MessageConsumedRecord) error
	GetMessages(topic string, group string) ([]*Message, error)
}

func NewRepository(db *gorm.DB) Repository {
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

func (r *repository) GetMessages(topic string, group string) ([]*Message, error) {
	var lastConsumedRecord MessageConsumedRecord
	if err := r.db.Where("topic = ? AND group = ?", topic, group).Last(&lastConsumedRecord).Error; err != nil {
		return nil, err
	}

	var messages []*Message
	if err := r.db.Where("topic = ? AND published_at >= ?", topic, lastConsumedRecord.ConsumedAt).Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}
