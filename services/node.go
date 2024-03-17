package services

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"geo-distributed-message-broker/models"
	"geo-distributed-message-broker/pb"
	"log/slog"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Node interface {
	Close() error
	Propose(req models.ProposeRequest) (models.ProposeResponse, error)
	Stable(req models.StableRequest) error
}

func NewNode(host string) (Node, error) {
	slog.Info("Creating new node client 📡", "node", host)

	// read ca's cert
	caCert, err := os.ReadFile("cert/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	// create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		return nil, err
	}

	//read client cert
	clientCert, err := tls.LoadX509KeyPair("cert/client-cert.pem", "cert/client-key.pem")
	if err != nil {
		return nil, err
	}

	// set config of tls credential
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	tlsCredential := credentials.NewTLS(config)

	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(tlsCredential))
	if err != nil {
		return nil, err
	}

	client := pb.NewNodeClient(conn)

	n := node{
		conn:   conn,
		client: client,
	}

	return &n, nil
}

type node struct {
	conn   *grpc.ClientConn
	client pb.NodeClient
}

func (n *node) Close() error {
	return n.conn.Close()
}

func (n *node) Propose(req models.ProposeRequest) (models.ProposeResponse, error) {
	rsp, err := n.client.Propose(context.TODO(), req.ToPb())
	if err != nil {
		return models.ProposeResponse{}, err
	}

	return models.ToProposeResponse(rsp), nil
}

func (n *node) Stable(req models.StableRequest) error {
	_, err := n.client.Stable(context.TODO(), req.ToPb())
	if err != nil {
		return err
	}

	return nil
}
