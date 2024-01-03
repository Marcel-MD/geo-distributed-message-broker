package services

import (
	"context"
	"geo-distributed-message-broker/models"
	"geo-distributed-message-broker/proto"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node interface {
	Close() error
	Propose(req models.ProposeRequest) (models.ProposeResponse, error)
	Stable(req models.StableRequest) error
}

func NewNode(host string) (Node, error) {
	slog.Info("Creating new node client 📡", "node", host)

	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := proto.NewNodeClient(conn)

	n := node{
		conn:   conn,
		client: client,
	}

	return &n, nil
}

type node struct {
	conn   *grpc.ClientConn
	client proto.NodeClient
}

func (n *node) Close() error {
	return n.conn.Close()
}

func (n *node) Propose(req models.ProposeRequest) (models.ProposeResponse, error) {
	rsp, err := n.client.Propose(context.TODO(), req.ToProto())
	if err != nil {
		return models.ProposeResponse{}, err
	}

	return models.ToProposeResponse(rsp), nil
}

func (n *node) Stable(req models.StableRequest) error {
	_, err := n.client.Stable(context.TODO(), req.ToProto())
	if err != nil {
		return err
	}

	return nil
}
