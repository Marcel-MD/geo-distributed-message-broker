package services

import (
	"context"
	"geo-distributed-message-broker/data"
	"log/slog"
	"sync"
	"time"
)

type Topic interface {
	GetMessages() map[string]data.Message
	AddMessage(msg data.Message, predecessors map[string]data.Message)
	RemoveMessage(id string)
	Wait(predecessors map[string]data.Message)
}

func NewTopic(name string) Topic {
	slog.Info("Creating new topic 🗃️", "name", name)

	return &topic{
		name:     name,
		messages: make(map[string]MessageTuple),
	}
}

type topic struct {
	name     string
	messages map[string]MessageTuple // map[message_id]MessageTuple
	mu       sync.RWMutex            // protects messages
}

type MessageTuple struct {
	message      data.Message
	predecessors map[string]data.Message // map[message_id]data.Message
	waitChannels []chan bool
}

func (t *topic) GetMessages() map[string]data.Message {
	t.mu.RLock()
	defer t.mu.RUnlock()

	messages := make(map[string]data.Message)
	for id, tuple := range t.messages {
		messages[id] = tuple.message
	}

	return messages
}

func (t *topic) AddMessage(msg data.Message, predecessors map[string]data.Message) {
	t.mu.Lock()
	t.messages[msg.ID] = MessageTuple{
		message:      msg,
		predecessors: predecessors,
		waitChannels: []chan bool{t.scheduleRemoveMessage(msg.ID)},
	}
	t.mu.Unlock()
}

func (t *topic) scheduleRemoveMessage(id string) chan bool {
	timeoutChan := make(chan bool, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		select {
		case <-ctx.Done():
			slog.Warn("Message timed out before being stable", "id", id)
			t.RemoveMessage(id)
		case <-timeoutChan:
		}
	}()

	return timeoutChan
}

func (t *topic) RemoveMessage(id string) {
	t.mu.Lock()
	if tuple, ok := t.messages[id]; ok {
		for _, w := range tuple.waitChannels {
			w <- true
		}
		delete(t.messages, id)
	}
	t.mu.Unlock()
}

func (t *topic) Wait(predecessors map[string]data.Message) {
	if len(predecessors) == 0 {
		return
	}

	var wg sync.WaitGroup

	t.mu.Lock()
	for _, predecessor := range predecessors {
		if tuple, ok := t.messages[predecessor.ID]; ok {
			wg.Add(1)
			waitChan := make(chan bool, 1)
			tuple.waitChannels = append(tuple.waitChannels, waitChan)
			go func() {
				<-waitChan
				wg.Done()
			}()
			t.messages[predecessor.ID] = tuple
		}
	}
	t.mu.Unlock()

	wg.Wait()
}
