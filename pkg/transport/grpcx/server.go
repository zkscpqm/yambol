package grpcx

import (
	"context"
	"fmt"
	"net"
	"time"
	"yambol/pkg/util"

	"yambol/pkg/broker"
	"yambol/pkg/transport/proto/grpcAPI"

	"google.golang.org/grpc"
)

type YambolGRPCServer struct {
	svr       *grpc.Server
	b         *broker.MessageBroker
	startedAt time.Time
	grpcAPI.APIServer
}

func NewYambolGRPCServer(b *broker.MessageBroker) *YambolGRPCServer {
	grpcServer := grpc.NewServer()

	//rtr.Use(LoggingMiddleware)

	svr := &YambolGRPCServer{
		svr:       grpcServer,
		b:         b,
		startedAt: time.Now(),
	}
	grpcAPI.RegisterAPIServer(grpcServer, svr)
	return svr
}

func (s *YambolGRPCServer) Serve(host string, port int) error {
	if s.svr == nil {
		return fmt.Errorf("cannot start server, server is nil")
	}
	target := fmt.Sprintf("%s:%d", host, port)
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

func (s *YambolGRPCServer) ServeLocal(port int) error {
	return s.Serve("localhost", port)
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
