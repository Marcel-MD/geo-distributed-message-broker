package services

import (
	"geo-distributed-message-broker/data"
	"log/slog"
	"sync"
	"time"
)

const MESSAGE_TTL = 10 * time.Second
const MESSAGE_CLEANUP = 20 * time.Second

type Messages map[string]data.Message // map[message_id]data.Message

type Topic interface {
	GetMessages(states ...string) Messages
	UpsertMessage(msg data.Message, state string, predecessors Messages) bool
	WaitForStateUpdate(predecessors Messages, states ...string) Messages
}

func NewTopic(name string) Topic {
	slog.Info("Creating new topic 🗃️", "name", name)

	t := &topic{
		name:     name,
		messages: make(map[string]MessageTuple),
	}
	time.AfterFunc(MESSAGE_CLEANUP, t.CleanupJob)

	return t
}

type topic struct {
	name     string
	messages map[string]MessageTuple // map[message_id]MessageTuple
	mu       sync.RWMutex            // protects messages
}

const (
	ProposedState  = "proposed"
	AckState       = "acknowledged"
	NackState      = "not_acknowledged"
	StableState    = "stable"
	PublishedState = "published"
	ExpiredState   = "expired"
)

type MessageTuple struct {
	message      data.Message
	predecessors Messages
	waitChannels []chan WaitResult
	state        string
	expire       int64
}

type WaitResult struct {
	State        string
	Predecessors Messages
}

func (t *MessageTuple) createWaitChannel() chan WaitResult {
	waitChan := make(chan WaitResult, 5)
	t.waitChannels = append(t.waitChannels, waitChan)
	return waitChan
}

func (t *MessageTuple) broadcastWaitResult() {
	if len(t.waitChannels) == 0 {
		return
	}

	result := WaitResult{
		State:        t.state,
		Predecessors: t.predecessors,
	}

	for _, w := range t.waitChannels {
		w <- result
	}
}

func (t *topic) CleanupJob() {
	messagesRemoved := 0

	t.mu.Lock()
	now := time.Now().Unix()
	for id, tuple := range t.messages {
		if tuple.expire < now {
			if tuple.state != PublishedState {
				slog.Warn("Message expired before being published", "topic", t.name, "message", id)
			}

			tuple.state = ExpiredState
			tuple.broadcastWaitResult()
			delete(t.messages, id)
			messagesRemoved++
		}
	}
	t.mu.Unlock()

	if messagesRemoved > 0 {
		slog.Info("Cleanup job finished", "topic", t.name, "messages_removed", messagesRemoved)
	}

	time.AfterFunc(MESSAGE_CLEANUP, t.CleanupJob)
}

func (t *topic) GetMessages(states ...string) Messages {
	if len(states) == 0 {
		return make(Messages)
	}

	statesMap := make(map[string]bool)
	for _, state := range states {
		statesMap[state] = true
	}

	t.mu.Lock()
	messages := make(Messages, len(t.messages))
	for id, tuple := range t.messages {
		if _, ok := statesMap[tuple.state]; ok {
			messages[id] = tuple.message
		}
	}
	t.mu.Unlock()

	return messages
}

func (t *topic) UpsertMessage(msg data.Message, state string, predecessors Messages) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	tuple, ok := t.messages[msg.ID]
	if !ok {
		// If no message exists, create a new one
		tuple.message = msg
		tuple.state = state
		tuple.predecessors = predecessors
		tuple.waitChannels = []chan WaitResult{}
		tuple.expire = time.Now().Add(MESSAGE_TTL).Unix()
		t.messages[msg.ID] = tuple
		return true
	}

	switch tuple.state {
	case PublishedState:
		return false
	case StableState:
		if state != PublishedState {
			return false
		}
	}

	if state == ProposedState {
		tuple.state = NackState
		tuple.broadcastWaitResult()

		tuple.message = msg
		tuple.state = state
		tuple.predecessors = predecessors
		tuple.waitChannels = []chan WaitResult{}
		tuple.expire = time.Now().Add(MESSAGE_TTL).Unix()
		t.messages[msg.ID] = tuple
		return true
	}

	tuple.state = state
	tuple.predecessors = predecessors
	tuple.broadcastWaitResult()
	t.messages[msg.ID] = tuple
	return false
}

func (t *topic) WaitForStateUpdate(messages Messages, states ...string) Messages {
	if len(messages) == 0 {
		return messages
	}

	endStatesMap := map[string]bool{
		NackState:      true,
		ExpiredState:   true,
		PublishedState: true,
	}

	desiredStatesMap := make(map[string]bool)
	for _, state := range states {
		desiredStatesMap[state] = true
	}

	predecessors := make(Messages)
	waitResultsChan := make(chan WaitResult, len(messages))
	var wg sync.WaitGroup

	t.mu.Lock()
	for _, msg := range messages {
		if tuple, ok := t.messages[msg.ID]; ok {
			if _, ok := endStatesMap[tuple.state]; ok {
				continue
			}

			if _, ok := desiredStatesMap[tuple.state]; ok {
				for id, msg := range tuple.predecessors {
					predecessors[id] = msg
				}
				continue
			}

			wg.Add(1)
			waitChan := tuple.createWaitChannel()
			go func() {
				for result := range waitChan {
					if _, ok := endStatesMap[result.State]; ok {
						wg.Done()
						return
					}

					if _, ok := desiredStatesMap[result.State]; ok {
						waitResultsChan <- result
						wg.Done()
						return
					}
				}
			}()
			t.messages[msg.ID] = tuple
		}
	}
	t.mu.Unlock()

	wg.Wait()
	close(waitResultsChan)
	for result := range waitResultsChan {
		for id, msg := range result.Predecessors {
			predecessors[id] = msg
		}
	}

	return predecessors
}
