package api

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"geo-distributed-message-broker/config"
	"geo-distributed-message-broker/models"
	"geo-distributed-message-broker/pb"
	"geo-distributed-message-broker/services"
	"log/slog"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func NewNodeServer(cfg config.Config, consensus services.ConsensusService) (*grpc.Server, net.Listener, error) {
	slog.Info("Creating new node server 🌐")

	// read ca's cert, verify to client's certificate
	caPem, err := os.ReadFile("cert/ca-cert.pem")
	if err != nil {
		return nil, nil, err
	}

	// create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPem) {
		return nil, nil, errors.New("failed to append ca's cert")
	}

	// read server cert & key
	serverCert, err := tls.LoadX509KeyPair("cert/server-cert.pem", "cert/server-key.pem")
	if err != nil {
		return nil, nil, err
	}

	// configuration of the certificate what we want to
	conf := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	// create tls credentials
	tlsCredentials := credentials.NewTLS(conf)

	srv := &nodeServer{
		consensus: consensus,
	}

	listener, err := net.Listen("tcp", cfg.NodePort)
	if err != nil {
		return nil, nil, err
	}

	grpcSrv := grpc.NewServer(grpc.Creds(tlsCredentials))

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
