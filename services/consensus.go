package services

import (
	"geo-distributed-message-broker/data"
	"geo-distributed-message-broker/models"
)

type ConsensusService interface {
	Publish(msg data.Message) (uint64, error)
	Propose(req models.ProposeRequest) (models.ProposeResponse, error)
	Stable(req models.StableRequest) error
}
