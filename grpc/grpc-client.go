package grpc

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcClient struct {
	conn *grpc.ClientConn
}

type GrpcClient interface {
	GetGrpcConnection() *grpc.ClientConn
	Close() error
}

type GrpcConfig struct {
	Port        string `json:"port"`
	Host        string `json:"host"`
	Development bool   `json:"development"`
}

func NewGrpcClient(config *GrpcConfig) (GrpcClient, error) {
	conn, err := grpc.NewClient(fmt.Sprintf("%s%s", config.Host, config.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &grpcClient{conn: conn}, nil
}

func (g *grpcClient) GetGrpcConnection() *grpc.ClientConn {
	return g.conn
}

func (g *grpcClient) Close() error {
	return g.conn.Close()
}
