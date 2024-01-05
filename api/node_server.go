package api

import (
	"context"
	"geo-distributed-message-broker/config"
	"geo-distributed-message-broker/models"
	"geo-distributed-message-broker/pb"
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

	pb.RegisterNodeServer(grpcSrv, srv)

	return grpcSrv, listener, nil
}

type nodeServer struct {
	pb.UnsafeNodeServer
	consensus services.ConsensusService
}

func (s *nodeServer) Propose(ctx context.Context, req *pb.ProposeRequest) (*pb.ProposeResponse, error) {
	rsp, err := s.consensus.Propose(models.ToProposeRequest(req))
	if err != nil {
		return nil, err
	}

	return rsp.ToPb(), nil
}

func (s *nodeServer) Stable(ctx context.Context, req *pb.StableRequest) (*pb.StableResponse, error) {
	err := s.consensus.Stable(models.ToStableRequest(req))
	if err != nil {
		return nil, err
	}

	return &pb.StableResponse{
		Ack: true,
	}, nil
}
