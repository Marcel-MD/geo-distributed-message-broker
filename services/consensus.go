package services

import (
	"errors"
	"geo-distributed-message-broker/config"
	"geo-distributed-message-broker/data"
	"geo-distributed-message-broker/models"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ConsensusService interface {
	Publish(msg data.Message) (string, error)
	Propose(req models.ProposeRequest) (models.ProposeResponse, error)
	Stable(req models.StableRequest) error
}

func NewConsensusService(cfg config.Config, broker BrokerService) ConsensusService {
	slog.Info("Creating new consensus service 🏛️")
	nodes := make(map[string]Node)

	for _, nodeHost := range cfg.Nodes {
		node, err := NewNode(nodeHost)
		if err != nil {
			slog.Error("Failed to create node client", "node", nodeHost, "error", err)
			continue
		}

		nodes[nodeHost] = node
	}

	return &consensusService{
		nodes:  nodes,
		topics: make(map[string]Topic),
		broker: broker,
	}
}

type consensusService struct {
	nodes  map[string]Node  // map[node_host]Node
	topics map[string]Topic // map[topic_name]Topic
	mu     sync.RWMutex     // protects topics
	broker BrokerService
}

func (c *consensusService) Publish(msg data.Message) (string, error) {
	if len(c.nodes) == 0 {
		return c.broker.Publish(msg)
	}

	// Prepare propose message request
	body := msg.Body
	msg.Body = []byte{}
	msg.ID = uuid.NewString()
	msg.Timestamp = time.Now().UnixMicro()

	proposeReq := models.ProposeRequest{
		Message: msg,
	}

	var predecessors map[string]data.Message
	var success bool = false

	// Propose message with max 3 retries
	for i := 0; i < 3; i++ {
		// Propose message to self
		rsp, err := c.Propose(proposeReq)
		if err != nil {
			proposeReq.Message.Timestamp++
			slog.Error("Failed to self propose message, retrying...", "error", err)
			continue
		}

		predecessors = make(map[string]data.Message)
		highestTimestamp := proposeReq.Message.Timestamp

		for id, msg := range rsp.Predecessors {
			if msg.Timestamp > highestTimestamp {
				highestTimestamp = msg.Timestamp
			}
			predecessors[id] = msg
		}

		if !rsp.Ack {
			proposeReq.Message.Timestamp = highestTimestamp + 1
			slog.Warn("Failed to self propose message, retrying...")
			continue
		}

		quorum := len(c.nodes)/2 + 1
		acks := 1
		nacks := 0

		type response struct {
			proposeResponse models.ProposeResponse
			err             error
		}
		responseChan := make(chan response, len(c.nodes))

		// Propose message to all other nodes
		for host, node := range c.nodes {
			go func(host string, node Node) {
				rsp, err := node.Propose(proposeReq)
				if err != nil {
					slog.Error("Failed to propose message", "node", host, "error", err)
					responseChan <- response{
						proposeResponse: models.ProposeResponse{},
						err:             err,
					}
				}

				responseChan <- response{
					proposeResponse: rsp,
				}

			}(host, node)
		}

		// Wait for quorum
		for rsp := range responseChan {
			if rsp.err != nil {
				quorum -= 1
				continue
			}

			proposeRsp := rsp.proposeResponse

			if proposeRsp.Ack {
				acks++
			} else {
				nacks++
			}

			for id, msg := range proposeRsp.Predecessors {
				if msg.Timestamp > highestTimestamp {
					highestTimestamp = msg.Timestamp
				}
				predecessors[id] = msg
			}

			if acks >= quorum || nacks >= quorum {
				break
			}
		}

		if acks < quorum {
			proposeReq.Message.Timestamp = highestTimestamp + 1
			slog.Warn("Failed to propose message to other nodes, retrying...")
			continue
		}

		success = true
		break
	}

	if !success {
		return "", errors.New("failed to propose message")
	}

	// Prepare stable message request
	msg.Body = body
	stableReq := models.StableRequest{
		Message:      msg,
		Predecessors: predecessors,
	}

	// Stable message to all other nodes
	for host, node := range c.nodes {
		go func(host string, node Node) {
			err := node.Stable(stableReq)
			if err != nil {
				slog.Error("Failed to send stable message", "node", host, "error", err)
			}
		}(host, node)
	}

	// Stable message to self
	err := c.Stable(stableReq)
	if err != nil {
		return "", err
	}

	return msg.ID, nil
}

func (c *consensusService) Propose(req models.ProposeRequest) (models.ProposeResponse, error) {
	slog.Debug("Receiving propose request", "message", req.Message.ID)

	// Create topic if it does not exist
	c.mu.Lock()
	topic := c.topics[req.Message.Topic]
	if topic == nil {
		topic = NewTopic(req.Message.Topic)
		c.topics[req.Message.Topic] = topic
	}
	c.mu.Unlock()

	// Remove older version of message if it exists and add new message
	topic.RemoveMessage(req.Message.ID)
	topic.AddMessage(req.Message, make(Messages))

	// Get all ack messages, split into newer and older messages based on timestamp
	ackMessages := topic.GetMessages(AckState)
	newerMessages := make(map[string]data.Message)
	olderMessages := make(map[string]data.Message)
	for _, m := range ackMessages {
		if m.Timestamp > req.Message.Timestamp {
			newerMessages[m.ID] = m
		} else {
			olderMessages[m.ID] = m
		}
	}

	// Wait for newer messages to be stable
	ack := true
	if len(newerMessages) > 0 {
		predecessors := topic.Wait(newerMessages)
		if _, ok := predecessors[req.Message.ID]; !ok {
			ack = false
		}
	}

	if ack {
		// Update predecessors and state
		ackMessages = olderMessages
		topic.UpdateMessage(req.Message.ID, AckState, ackMessages)
		slog.Debug("Propose request acknowledged", "message", req.Message.ID)
	} else {
		// Remove message from topic
		topic.RemoveMessage(req.Message.ID)
		slog.Debug("Propose request not acknowledged", "message", req.Message.ID)
	}

	return models.ProposeResponse{
		Ack:          ack,
		Message:      req.Message,
		Predecessors: ackMessages,
	}, nil
}

func (c *consensusService) Stable(req models.StableRequest) error {
	slog.Debug("Receiving stable request", "message", req.Message.ID)

	// Create topic if it does not exist
	c.mu.Lock()
	topic := c.topics[req.Message.Topic]
	if topic == nil {
		topic = NewTopic(req.Message.Topic)
		c.topics[req.Message.Topic] = topic
	}
	c.mu.Unlock()

	// Update predecessors and state
	topic.UpdateMessage(req.Message.ID, StableState, req.Predecessors)

	// Remove message from topic
	topic.RemoveMessage(req.Message.ID)

	// Wait for predecessors to be stable
	topic.Wait(req.Predecessors)

	// Publish message to broker
	_, err := c.broker.Publish(req.Message)
	if err != nil {
		return err
	}

	return nil
}
