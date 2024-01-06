package services

import (
	"context"
	"geo-distributed-message-broker/data"
	"log/slog"
	"sync"
	"time"
)

type Messages map[string]data.Message // map[message_id]data.Message

type Topic interface {
	GetMessages(state string) Messages
	AddMessage(msg data.Message, predecessors Messages)
	RemoveMessage(id string)
	UpdateMessage(id string, state string, predecessors Messages)
	Wait(predecessors Messages) Messages
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

const (
	ProposedState = "proposed"
	AckState      = "acknowledged"
	NackState     = "not acknowledged"
	StableState   = "stable"
)

type MessageTuple struct {
	message      data.Message
	predecessors Messages
	waitChannels []chan Messages
	state        string
}

func (t *topic) GetMessages(state string) Messages {
	t.mu.RLock()
	defer t.mu.RUnlock()

	messages := make(Messages, len(t.messages))
	for id, tuple := range t.messages {
		if tuple.state == state {
			messages[id] = tuple.message
		}
	}

	return messages
}

func (t *topic) AddMessage(msg data.Message, predecessors Messages) {
	t.mu.Lock()
	t.messages[msg.ID] = MessageTuple{
		message:      msg,
		predecessors: predecessors,
		waitChannels: []chan Messages{t.scheduleMessageTimeout(msg.ID)},
		state:        ProposedState,
	}
	t.mu.Unlock()
}

func (t *topic) scheduleMessageTimeout(id string) chan Messages {
	timeoutChan := make(chan Messages, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
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
			w <- tuple.predecessors
		}
		delete(t.messages, id)
	}
	t.mu.Unlock()
}

func (t *topic) UpdateMessage(id string, state string, predecessors Messages) {
	t.mu.Lock()
	if tuple, ok := t.messages[id]; ok {
		tuple.state = state
		tuple.predecessors = predecessors
		t.messages[id] = tuple
	}
	t.mu.Unlock()
}

func (t *topic) Wait(messages Messages) Messages {
	if len(messages) == 0 {
		return messages
	}

	predecessors := make(Messages)
	predecessorsChan := make(chan Messages, len(messages))
	var wg sync.WaitGroup

	t.mu.Lock()
	for _, msg := range messages {
		if tuple, ok := t.messages[msg.ID]; ok {
			wg.Add(1)
			waitChan := make(chan Messages, 1)
			tuple.waitChannels = append(tuple.waitChannels, waitChan)
			go func() {
				predecessorsChan <- <-waitChan
				wg.Done()
			}()
			t.messages[msg.ID] = tuple
		}
	}
	t.mu.Unlock()

	wg.Wait()
	close(predecessorsChan)
	for predecessor := range predecessorsChan {
		for id, msg := range predecessor {
			predecessors[id] = msg
		}
	}

	return predecessors
}
