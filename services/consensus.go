package services

import (
	"encoding/json"
	"errors"
	"geo-distributed-message-broker/config"
	"geo-distributed-message-broker/data"
	"log/slog"
	"sync"
	"time"

	anotherSlog "github.com/gookit/slog"

	"github.com/google/uuid"
	"github.com/madalv/conalg/caesar"
)

type ConsensusService interface {
	Publish(msg data.Message) (string, error)
	DetermineConflict(c1, c2 []byte) bool
	Execute(c []byte)
}

func NewConsensusService(cfg config.Config, broker BrokerService) ConsensusService {
	slog.Info("Creating new consensus service 🏛️")

	srv := consensusService{
		proposedMessages: make(map[string]chan messageResponse),
		cfg:              cfg,
		broker:           broker,
	}

	conalg := caesar.InitConalgModule(&srv, "", anotherSlog.InfoLevel, true)
	srv.conalg = conalg

	return &srv
}

type messageResponse struct {
	Message data.Message
	Err     error
}

type consensusService struct {
	proposedMessages map[string]chan messageResponse
	mx               sync.RWMutex
	cfg              config.Config
	conalg           caesar.Conalg
	broker           BrokerService
}

func (c *consensusService) DetermineConflict(message1, message2 []byte) bool {
	var msg1 data.Message
	err := json.Unmarshal(message1, &msg1)
	if err != nil {
		slog.Error("Failed to unmarshal message1", "error", err)
		return true
	}

	var msg2 data.Message
	err = json.Unmarshal(message2, &msg2)
	if err != nil {
		slog.Error("Failed to unmarshal message2", "error", err)
		return true
	}

	if msg1.Topic == msg2.Topic {
		return true
	}

	return false
}

func (c *consensusService) Execute(message []byte) {
	var msg data.Message
	err := json.Unmarshal(message, &msg)
	if err != nil {
		slog.Error("Failed to unmarshal message", "error", err)
		return
	}

	rsp := messageResponse{
		Message: msg,
	}

	_, err = c.broker.Publish(msg)
	if err != nil {
		rsp.Err = err
	}

	c.mx.Lock()
	if ch, ok := c.proposedMessages[msg.ID]; ok {
		ch <- rsp
		delete(c.proposedMessages, msg.ID)
	}
	c.mx.Unlock()
}

func (c *consensusService) Publish(msg data.Message) (string, error) {
	if len(c.cfg.Nodes) == 0 {
		return c.broker.Publish(msg)
	}

	msg.ID = uuid.NewString()
	msg.Timestamp = time.Now().UnixMicro()

	message, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	waitChan := make(chan messageResponse, 2)
	c.mx.Lock()
	c.proposedMessages[msg.ID] = waitChan
	c.mx.Unlock()
	c.conalg.Propose(message)

	return wait(waitChan)
}

func wait(waitChan chan messageResponse) (string, error) {
	select {
	case msg := <-waitChan:
		return msg.Message.ID, msg.Err
	case <-time.After(5 * time.Second):
		return "", errors.New("timeout waiting for consensus")
	}
}
