package data

import "time"

type Message struct {
	ID          string    `json:"id" gorm:"primaryKey;autoIncrement:false"`
	Topic       string    `json:"topic" gorm:"primaryKey"`
	Body        []byte    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
}

type MessageConsumedRecord struct {
	MessageID  string    `json:"message_id" gorm:"primaryKey"`
	Topic      string    `json:"topic" gorm:"primaryKey"`
	ConsumedBy string    `json:"consumed_by" gorm:"primaryKey"`
	ConsumedAt time.Time `json:"consumed_at"`
}
