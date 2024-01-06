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
	GetAckMessages() Messages
	AddAckMessage(msg data.Message, predecessors Messages)
	RemoveAckMessage(id string)
	UpdatePredecessors(id string, predecessors Messages)
	Wait(predecessors Messages) Messages
}

func NewTopic(name string) Topic {
	slog.Info("Creating new topic 🗃️", "name", name)

	return &topic{
		name:        name,
		ackMessages: make(map[string]MessageTuple),
	}
}

type topic struct {
	name        string
	ackMessages map[string]MessageTuple // map[message_id]MessageTuple
	mu          sync.RWMutex            // protects messages
}

type MessageTuple struct {
	message      data.Message
	predecessors Messages
	waitChannels []chan Messages
}

func (t *topic) GetAckMessages() Messages {
	t.mu.RLock()
	defer t.mu.RUnlock()

	messages := make(Messages, len(t.ackMessages))
	for id, tuple := range t.ackMessages {
		messages[id] = tuple.message
	}

	return messages
}

func (t *topic) AddAckMessage(msg data.Message, predecessors Messages) {
	t.mu.Lock()
	t.ackMessages[msg.ID] = MessageTuple{
		message:      msg,
		predecessors: predecessors,
		waitChannels: []chan Messages{t.scheduleRemoveMessage(msg.ID)},
	}
	t.mu.Unlock()
}

func (t *topic) scheduleRemoveMessage(id string) chan Messages {
	timeoutChan := make(chan Messages, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		select {
		case <-ctx.Done():
			slog.Warn("Message timed out before being stable", "id", id)
			t.RemoveAckMessage(id)
		case <-timeoutChan:
		}
	}()

	return timeoutChan
}

func (t *topic) RemoveAckMessage(id string) {
	t.mu.Lock()
	if tuple, ok := t.ackMessages[id]; ok {
		for _, w := range tuple.waitChannels {
			w <- tuple.predecessors
		}
		delete(t.ackMessages, id)
	}
	t.mu.Unlock()
}

func (t *topic) UpdatePredecessors(id string, predecessors Messages) {
	t.mu.Lock()
	if tuple, ok := t.ackMessages[id]; ok {
		tuple.predecessors = predecessors
		t.ackMessages[id] = tuple
	}
	t.mu.Unlock()
}

func (t *topic) Wait(predecessors Messages) Messages {
	if len(predecessors) == 0 {
		return predecessors
	}

	predecessorsOfPredecessors := make(Messages)
	predecessorsChan := make(chan Messages, len(predecessors))
	var wg sync.WaitGroup

	t.mu.Lock()
	for _, predecessor := range predecessors {
		if tuple, ok := t.ackMessages[predecessor.ID]; ok {
			wg.Add(1)
			waitChan := make(chan Messages, 1)
			tuple.waitChannels = append(tuple.waitChannels, waitChan)
			go func() {
				predecessorsChan <- <-waitChan
				wg.Done()
			}()
			t.ackMessages[predecessor.ID] = tuple
		}
	}
	t.mu.Unlock()

	wg.Wait()
	close(predecessorsChan)
	for predecessor := range predecessorsChan {
		for id, msg := range predecessor {
			predecessorsOfPredecessors[id] = msg
		}
	}

	return predecessorsOfPredecessors
}
