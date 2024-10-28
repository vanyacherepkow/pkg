package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

const (
	maxConnectionIdle = 5
	gRPCTimeout       = 15
	maxConnectionAge  = 5
	gRPCTime          = 10
)

type GrpcServer struct {
	Grpc   *grpc.Server
	Config *GrpcConfig
	//TODO: Create logger
}

func NewGrpcServer(config *GrpcConfig) *GrpcServer {
	//TODO: Add otelgrpc

	s := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: maxConnectionIdle * time.Minute,
			Timeout:           gRPCTimeout * time.Second,
			MaxConnectionAge:  maxConnectionAge * time.Minute,
			Time:              gRPCTime * time.Minute,
		}),
		//TODO: middleware for otelgrpc
	)

	return &GrpcServer{Grpc: s, Config: config}
}

func (s *GrpcServer) RunGrpcServer(ctx context.Context, configGrpc ...func(grpcServer *grpc.Server)) error {
	fmt.Println(s.Config)
	fmt.Println(s.Config.Port)
	listen, err := net.Listen("tcp", s.Config.Port)
	if err != nil {
		return errors.Wrap(err, "net.Listen")
	}

	if len(configGrpc) > 0 {
		grpcFunc := configGrpc[0]
		if grpcFunc != nil {
			grpcFunc(s.Grpc)
		}
	}

	if s.Config.Development {
		reflection.Register(s.Grpc)
	}

	if len(configGrpc) > 0 {
		grpcFunc := configGrpc[0]
		if grpcFunc != nil {
			grpcFunc(s.Grpc)
		}
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("Shutting down grpc PORT: {%s}", s.Config.Port)
				s.shutdown()
				fmt.Printf("Grpc stopped")
				return
			}
		}
	}()

	fmt.Printf("gRPC Server is listening on port: %s", s.Config.Port)

	err = s.Grpc.Serve(listen)

	if err != nil {
		fmt.Println(fmt.Sprintf("[grpcServer_RunGrpcServer.Serve] grpc server serve error: %+v", err))
	}

	return nil
}

func (s *GrpcServer) shutdown() {
	s.Grpc.Stop()
	s.Grpc.GracefulStop()
}
