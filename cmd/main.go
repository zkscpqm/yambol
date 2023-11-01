package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"yambol/config"
	"yambol/pkg/broker"
	"yambol/pkg/transport/grpcx"
	"yambol/pkg/transport/httpx/rest"
	"yambol/pkg/util"
	"yambol/pkg/util/log"
)

const (
	DefaultRESTPortInsecure = 21419
	DefaultRESTPortSecure   = 21420
	DefaultGRPCPortInsecure = 21421
	DefaultGRPCPortSecure   = 21422
)

func main() {

	cfg, err := config.FromFile()
	if err != nil {
		fmt.Printf("failed to load config file: %v\n", err)
		os.Exit(1)
	}

	fh, err := log.NewDefaultFileHandler(cfg.Log.File)
	if err != nil {
		fmt.Println("failed to create file handler:", err)
		os.Exit(1)
	}

	logger := log.New("MAIN", log.ParseLevel(cfg.Log.Level), fh, log.NewDefaultStdioHandler())
	logger.Info("---------------------------------Running Yambol---------------------------------")
	logger.Info("Initializing config manager...")
	config.Init(*cfg, logger)

	logger.Info("Setting defaults...")
	broker.SetDefaultMinLen(cfg.Broker.DefaultMinLength)
	broker.SetDefaultMaxLen(cfg.Broker.DefaultMaxLength)
	broker.SetDefaultMaxSizeBytes(cfg.Broker.DefaultMaxSizeBytes)
	broker.SetDefaultTTL(util.Seconds(cfg.Broker.DefaultTTL))

	b := broker.New(logger)
	for qName, qCfg := range cfg.Broker.Queues {
		if err = b.AddQueue(qName, broker.QueueOptions{
			MinLen:       qCfg.MinLength,
			MaxLen:       qCfg.MaxLength,
			MaxSizeBytes: qCfg.MaxSizeBytes,
			DefaultTTL:   util.Seconds(qCfg.TTL),
		}); err != nil {
			logger.Error("failed to add queue: %v", err)
		}
	}

	certPath, err := filepath.Abs(cfg.API.Certificate)
	if err != nil {
		logger.Error("failed to get TLS certificate path: %v", err)
	}
	keyPath, err := filepath.Abs(cfg.API.Key)
	if err != nil {
		logger.Error("failed to get TLS key path: %v", err)
	}

	var wg sync.WaitGroup

	runRESTServer := func() {
		if !cfg.API.REST.Enabled {
			return
		}
		wg.Add(1)
		s := rest.NewServer(b, nil, logger)
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
					logger.Error("REST (tls=on) server crashed: %v", err)
				}
				wg.Done()
			}()
		} else {
			go func() {
				err = s.ListenAndServeInsecure(port)
				if err != nil {
					logger.Error("REST (tls=off) server crashed: %v", err)
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
			logger.Error("failed to create gRPC server: %v", err)
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
				logger.Error("GRPC server crashed: %v", err)
			}
			wg.Done()
		}()
	}

	runRESTServer()
	runGRPCServer()
	wg.Wait()
}
