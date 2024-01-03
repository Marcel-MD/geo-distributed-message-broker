package services

import (
	"errors"
	"geo-distributed-message-broker/config"
	"geo-distributed-message-broker/data"
	"geo-distributed-message-broker/models"
	"log/slog"
)

type ConsensusService interface {
	Publish(msg data.Message) (uint64, error)
	Propose(req models.ProposeRequest) (models.ProposeResponse, error)
	Stable(req models.StableRequest) error
}

func NewConsensusService(cfg config.Config) ConsensusService {
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
		nodes: nodes,
	}
}

type consensusService struct {
	nodes map[string]Node
}

func (c *consensusService) Publish(msg data.Message) (uint64, error) {
	return 0, errors.New("not implemented")
}

func (c *consensusService) Propose(req models.ProposeRequest) (models.ProposeResponse, error) {
	return models.ProposeResponse{}, errors.New("not implemented")
}

func (c *consensusService) Stable(req models.StableRequest) error {
	return errors.New("not implemented")
}
