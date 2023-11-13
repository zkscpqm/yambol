package grpcx

import (
	"context"
	"fmt"
	"time"

	"yambol/pkg/transport/model"
	"yambol/pkg/transport/proto/grpcAPI"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	md "google.golang.org/grpc/metadata"
)

type Client struct {
	c grpcAPI.APIClient
}

type MetadataHook func(header, trailer *md.MD)

func (c *Client) Home(ctx context.Context) (*model.BasicInfo, error) {
	resp, err := c.c.Home(ctx, &grpcAPI.HomeRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to call Home: %v", err)
	}
	uptime, err := time.ParseDuration(resp.GetUptime())
	if err != nil {
		return nil, fmt.Errorf("failed to parse uptime: %v", err)
	}
	return &model.BasicInfo{Uptime: uptime, Version: resp.GetVersion()}, nil
}

func dialOptions(idleCheck, timeout time.Duration, creds credentials.TransportCredentials) []grpc.DialOption {
	if creds == nil {
		creds = insecure.NewCredentials()
	}
	return []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                idleCheck,
			Timeout:             timeout,
			PermitWithoutStream: true,
		}),
		grpc.WithTransportCredentials(creds),
	}
}

func NewClient(host string, port int, creds credentials.TransportCredentials, idleConnCheckSeconds, timeoutMinutes int) (*Client, error) {
	connStr := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.Dial(connStr, dialOptions(
		time.Duration(idleConnCheckSeconds)*time.Second,
		time.Duration(timeoutMinutes)*time.Minute,
		creds,
	)...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial to server %s: %v", connStr, err)
	}
	return &Client{
		c: grpcAPI.NewAPIClient(conn),
	}, nil
}

func NewDefaultClient(host string, port int, creds credentials.TransportCredentials) (*Client, error) {
	return NewClient(host, port, creds, 30, 15)
}

func NewDefaultInsecureClient(host string, port int) (*Client, error) {
	return NewDefaultClient(host, port, insecure.NewCredentials())
}
