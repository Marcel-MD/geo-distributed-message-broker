package data

type Message struct {
	ID        string `json:"id" gorm:"primaryKey"`
	Timestamp int64  `json:"timestamp"`
	Topic     string `json:"topic"`
	Body      []byte `json:"body"`
}
