package api

import (
	"context"
	"geo-distributed-message-broker/config"
	"geo-distributed-message-broker/data"
	"geo-distributed-message-broker/pb"
	"geo-distributed-message-broker/services"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

func NewBrokerServer(cfg config.Config, broker services.BrokerService, consensus services.ConsensusService) (*grpc.Server, net.Listener, error) {
	slog.Info("Creating new broker server 🌐")

	srv := &brokerServer{
		broker:    broker,
		consensus: consensus,
	}

	listener, err := net.Listen("tcp", cfg.BrokerPort)
	if err != nil {
		return nil, nil, err
	}

	grpcSrv := grpc.NewServer()

	pb.RegisterBrokerServer(grpcSrv, srv)

	return grpcSrv, listener, nil
}

type brokerServer struct {
	pb.UnsafeBrokerServer
	broker    services.BrokerService
	consensus services.ConsensusService
}

func (s *brokerServer) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	msg := data.Message{
		Topic: req.Topic,
		Body:  req.Body,
	}

	id, err := s.consensus.Publish(msg)
	if err != nil {
		return nil, err
	}

	rsp := &pb.PublishResponse{
		Id: id,
	}

	return rsp, nil
}

func (s *brokerServer) Subscribe(req *pb.SubscribeRequest, srv pb.Broker_SubscribeServer) error {
	ch, err := s.broker.Subscribe(req.ConsumerName, req.Topics)
	if err != nil {
		slog.Error("Failed to subscribe", "error", err.Error())
		return err
	}

	for {
		select {
		case msg := <-ch:
			rsp := &pb.MessageResponse{
				Id:        msg.ID,
				Timestamp: msg.Timestamp,
				Topic:     msg.Topic,
				Body:      msg.Body,
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
