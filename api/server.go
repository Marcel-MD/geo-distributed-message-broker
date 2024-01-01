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

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewServer(cfg config.Config, broker domain.Broker) (*grpc.Server, net.Listener, error) {
	slog.Info("Creating new GRPC server 🌐")

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
		Topic:       req.Topic,
		Body:        req.Body,
		PublishedAt: time.Now(),
	}

	id, err := s.broker.Publish(msg)
	if err != nil {
		return nil, err
	}

	rsp := &proto.PublishResponse{
		Id:          id,
		PublishedAt: timestamppb.New(msg.PublishedAt),
	}

	return rsp, nil
}

func (s *server) Subscribe(req *proto.SubscribeRequest, srv proto.Broker_SubscribeServer) error {
	ch, err := s.broker.Subscribe(req.ConsumerName, req.Topics)
	if err != nil {
		slog.Error("Failed to subscribe", "error", err.Error())
		return err
	}

	for {
		select {
		case msg := <-ch:
			rsp := &proto.MessageResponse{
				Id:          msg.ID,
				Topic:       msg.Topic,
				Body:        msg.Body,
				PublishedAt: timestamppb.New(msg.PublishedAt),
			}

			if err := srv.Send(rsp); err != nil {
				slog.Error("Failed to send message", "error", err.Error())
				s.broker.Unsubscribe(req.ConsumerName, req.Topics)
				return err
			}

			if err := s.broker.Acknowledge(req.ConsumerName, msg); err != nil {
				slog.Error("Failed to acknowledge message", "error", err.Error())
				s.broker.Unsubscribe(req.ConsumerName, req.Topics)
				return err
			}

		case <-srv.Context().Done():
			s.broker.Unsubscribe(req.ConsumerName, req.Topics)
			return nil
		}
	}
}
