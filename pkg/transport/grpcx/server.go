package grpcx

import (
	"context"
	"fmt"
	"net"
	"time"

	"yambol/pkg/broker"
	"yambol/pkg/transport/proto/grpcAPI"
	"yambol/pkg/util"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type YambolGRPCServer struct {
	svr       *grpc.Server
	b         *broker.MessageBroker
	startedAt time.Time
	grpcAPI.APIServer
}

func NewYambolGRPCServer(b *broker.MessageBroker, tlsEnabled bool, certFile, keyFile string) (*YambolGRPCServer, error) {
	opts := make([]grpc.ServerOption, 0)
	if tlsEnabled {
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS Credentials")
		}
		opts = append(opts, grpc.Creds(creds))
	}
	grpcServer := grpc.NewServer(opts...)

	svr := &YambolGRPCServer{
		svr:       grpcServer,
		b:         b,
		startedAt: time.Now(),
	}
	grpcAPI.RegisterAPIServer(grpcServer, svr)
	return svr, nil
}

func (s *YambolGRPCServer) ListenAndServe(port int) error {
	if s.svr == nil {
		return fmt.Errorf("cannot start server, server is nil")
	}
	target := fmt.Sprintf("localhost:%d", port)
	s.startedAt = time.Now()
	lis, err := net.Listen("tcp", target)
	if err != nil {
		return fmt.Errorf("failed to listen on tcp %s...: %v", target, err)
	}
	err = s.svr.Serve(lis)
	if err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}

func (s *YambolGRPCServer) Close(force bool) {
	if force {
		s.svr.Stop()
	} else {
		s.svr.GracefulStop()
	}
}

func (s *YambolGRPCServer) Home(ctx context.Context, _ *grpcAPI.HomeRequest) (*grpcAPI.HomeResponse, error) {
	_ = extractPeerInfo(ctx)
	return &grpcAPI.HomeResponse{Version: util.Version(), Uptime: time.Since(s.startedAt).String()}, nil
}
