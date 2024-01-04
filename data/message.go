package data

type Message struct {
	ID        string `json:"id" gorm:"primaryKey"`
	Timestamp int64  `json:"timestamp"`
	Topic     string `json:"topic"`
	Body      []byte `json:"body"`
}

type MessageConsumedRecord struct {
	MessageID  string `json:"message_id" gorm:"primaryKey"`
	ConsumedBy string `json:"consumed_by" gorm:"primaryKey"`
	Timestamp  int64  `json:"timestamp"`
	Topic      string `json:"topic"`
}
