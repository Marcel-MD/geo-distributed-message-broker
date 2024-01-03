package data

type Message struct {
	ID    uint64 `json:"id" gorm:"primaryKey;autoIncrement:false"`
	Topic string `json:"topic" gorm:"primaryKey"`
	Body  []byte `json:"body"`
}

type MessageConsumedRecord struct {
	MessageID  uint64 `json:"message_id" gorm:"primaryKey"`
	Topic      string `json:"topic" gorm:"primaryKey"`
	ConsumedBy string `json:"consumed_by" gorm:"primaryKey"`
}
