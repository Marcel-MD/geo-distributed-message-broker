package api

import (
	"context"
	"geo-distributed-message-broker/api/proto"
	"geo-distributed-message-broker/config"
	"geo-distributed-message-broker/data"
	"geo-distributed-message-broker/domain"
	"log/slog"
	"net"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewServer(cfg config.Config, broker domain.Broker) (*grpc.Server, net.Listener, error) {
	slog.Info("Creating new GRPC server")

	srv := &server{
		broker: broker,
	}

	listener, err := net.Listen("tcp", cfg.Port)
	if err != nil {
		return nil, nil, err
	}

	grpcSrv := grpc.NewServer()

	proto.RegisterBrokerServer(grpcSrv, srv)

	return grpcSrv, listener, nil
}

type server struct {
	proto.UnsafeBrokerServer
	broker domain.Broker
}

func (s *server) Publish(ctx context.Context, req *proto.PublishRequest) (*proto.PublishResponse, error) {
	msg := data.Message{
		ID:          uuid.New().String(),
		Topic:       req.Topic,
		Body:        req.Body,
		PublishedAt: time.Now(),
	}

	err := s.broker.Publish(msg)
	if err != nil {
		return nil, err
	}

	rsp := &proto.PublishResponse{
		Id:          msg.ID,
		PublishedAt: timestamppb.New(msg.PublishedAt),
	}

	return rsp, nil
}

func (s *server) Subscribe(req *proto.SubscribeRequest, srv proto.Broker_SubscribeServer) error {
	ch, err := s.broker.Subscribe(req.Name, req.Topic)
	if err != nil {
		return err
	}

	for msg := range ch {
		rsp := &proto.MessageResponse{
			Id:          msg.ID,
			Topic:       msg.Topic,
			Body:        msg.Body,
			PublishedAt: timestamppb.New(msg.PublishedAt),
		}

		if err := srv.Send(rsp); err != nil {
			s.broker.Unsubscribe(req.Name, req.Topic)
			return err
		}

		if err := s.broker.Acknowledge(req.Name, msg); err != nil {
			s.broker.Unsubscribe(req.Name, req.Topic)
			return err
		}
	}

	return nil
}
