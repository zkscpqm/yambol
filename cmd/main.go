package main

import (
	"fmt"
	"path/filepath"
	"sync"

	"yambol/config"
	"yambol/pkg/broker"
	"yambol/pkg/transport/grpcx"
	"yambol/pkg/transport/httpx/rest"
	"yambol/pkg/util"
)

const (
	DefaultRESTPortInsecure = 21419
	DefaultRESTPortSecure   = 21420
	DefaultGRPCPortInsecure = 21421
	DefaultGRPCPortSecure   = 21422
)

func main() {
	fmt.Println("Running Yambol...")
	cfg, err := config.FromFile()
	if err != nil {
		fmt.Println("failed to load config file: ", err)
		return
	}
	config.Init(*cfg)

	broker.SetDefaultMinLen(cfg.Broker.DefaultMinLength)
	broker.SetDefaultMaxLen(cfg.Broker.DefaultMaxLength)
	broker.SetDefaultMaxSizeBytes(cfg.Broker.DefaultMaxSizeBytes)
	broker.SetDefaultTTL(util.Seconds(cfg.Broker.DefaultTTL))

	b, err := broker.New()
	if err != nil {
		fmt.Println("failed to create broker:", err)
		return
	}
	for qName, qCfg := range cfg.Broker.Queues {
		if err = b.AddQueue(qName, broker.QueueOptions{
			MinLen:       qCfg.MinLength,
			MaxLen:       qCfg.MaxLength,
			MaxSizeBytes: qCfg.MaxSizeBytes,
			DefaultTTL:   util.Seconds(qCfg.TTL),
		}); err != nil {
			fmt.Println("failed to add queue: ", err)
			return
		}
	}

	certPath, err := filepath.Abs(cfg.API.Certificate)
	if err != nil {
		fmt.Println("failed to get TLS certificate path: ", err)
		return
	}
	keyPath, err := filepath.Abs(cfg.API.Key)
	if err != nil {
		fmt.Println("failed to get TLS key path: ", err)
		return
	}

	var wg sync.WaitGroup

	runRESTServer := func() {
		if !cfg.API.REST.Enabled {
			return
		}
		wg.Add(1)
		s := rest.NewYambolRESTServer(b, nil)
		port := cfg.API.REST.Port
		if port <= 0 {
			if cfg.API.REST.TlsEnabled {
				port = DefaultRESTPortSecure
			} else {
				port = DefaultRESTPortInsecure
			}
		}
		if cfg.API.REST.TlsEnabled {
			go func() {
				err = s.ListenAndServe(port, certPath, keyPath)
				if err != nil {
					fmt.Println("REST (tls=on) server crashed", err)
				}
				wg.Done()
			}()
		} else {
			go func() {
				err = s.ListenAndServeInsecure(port)
				if err != nil {
					fmt.Println("REST (tls=off) server crashed", err)
				}
				wg.Done()
			}()
		}
	}

	runGRPCServer := func() {
		if !cfg.API.GRPC.Enabled {
			return
		}
		var s *grpcx.YambolGRPCServer
		s, err = grpcx.NewYambolGRPCServer(b, cfg.API.GRPC.TlsEnabled, certPath, keyPath)
		if err != nil {
			fmt.Println("failed to create gRPC server: ", err)
			return
		}
		wg.Add(1)

		port := cfg.API.GRPC.Port
		if port <= 0 {
			if cfg.API.GRPC.TlsEnabled {
				port = DefaultGRPCPortSecure
			} else {
				port = DefaultGRPCPortInsecure
			}
		}
		go func() {
			err = s.ListenAndServe(port)
			if err != nil {
				fmt.Println("GRPC server crashed:", err)
			}
			wg.Done()
		}()
	}

	runRESTServer()
	runGRPCServer()
	wg.Wait()
}
