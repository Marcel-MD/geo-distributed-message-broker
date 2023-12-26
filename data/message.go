package data

import "time"

type Message struct {
	ID          uint64    `json:"id" gorm:"primaryKey;autoIncrement:false"`
	Topic       string    `json:"topic" gorm:"primaryKey"`
	Body        []byte    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
}

type MessageConsumedRecord struct {
	MessageID  uint64    `json:"message_id" gorm:"primaryKey"`
	Topic      string    `json:"topic" gorm:"primaryKey"`
	Group      string    `json:"group" gorm:"primaryKey"`
	ConsumedAt time.Time `json:"consumed_at"`
}
