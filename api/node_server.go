package api

import (
	"context"
	"geo-distributed-message-broker/config"
	"geo-distributed-message-broker/models"
	"geo-distributed-message-broker/proto"
	"geo-distributed-message-broker/services"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

func NewNodeServer(cfg config.Config, consensus services.ConsensusService) (*grpc.Server, net.Listener, error) {
	slog.Info("Creating new node server 🌐")

	srv := &nodeServer{
		consensus: consensus,
	}

	listener, err := net.Listen("tcp", cfg.NodePort)
	if err != nil {
		return nil, nil, err
	}

	grpcSrv := grpc.NewServer()

	proto.RegisterNodeServer(grpcSrv, srv)

	return grpcSrv, listener, nil
}

type nodeServer struct {
	proto.UnsafeNodeServer
	consensus services.ConsensusService
}

func (s *nodeServer) Propose(ctx context.Context, req *proto.ProposeRequest) (*proto.ProposeResponse, error) {
	rsp, err := s.consensus.Propose(models.ToProposeRequest(req))
	if err != nil {
		return nil, err
	}

	return rsp.ToProto(), nil
}

func (s *nodeServer) Stable(ctx context.Context, req *proto.StableRequest) (*proto.StableResponse, error) {
	err := s.consensus.Stable(models.ToStableRequest(req))
	if err != nil {
		return nil, err
	}

	return &proto.StableResponse{
		Ack: true,
	}, nil
}
