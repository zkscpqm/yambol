package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"yambol/pkg/transport/model"

	"yambol/config"
	"yambol/pkg/transport/httpx"
	"yambol/pkg/util"
)

func (s *Server) home() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		return s.respond(w, httpx.HomeResponse{
			StatusCode: http.StatusOK,
			BasicInfo: model.BasicInfo{
				Uptime:  time.Since(s.startedAt),
				Version: util.Version(),
			},
		})
	}
}

func (s *Server) stats() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		return s.respond(w, httpx.StatsResponse(s.b.Stats()))
	}
}

func (s *Server) runningConfig() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		target, err := resolveHTTPMethodTarget(r, map[string]HandlerFunc{
			http.MethodGet:  s.getRunningConfig(),
			http.MethodPost: s.setRunningConfig(),
		})
		if err != nil {
			return s.error(w, http.StatusMethodNotAllowed, err, r.Method)
		}
		return target(w, r)
	}
}

func (s *Server) getRunningConfig() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		return s.respond(w, httpx.ConfigResponse{
			StatusCode: http.StatusOK,
			Config:     config.GetRunningConfig(),
		})
	}
}

func (s *Server) setRunningConfig() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		var (
			cfg config.Configuration
			err error
		)
		err = json.NewDecoder(r.Body).Decode(&cfg)
		if err != nil {
			return s.error(w, http.StatusBadRequest, fmt.Errorf("failed to decode request body: %v", err))
		}
		config.SetRunningConfig(cfg)
		return s.respond(w, httpx.EmptyResponse{StatusCode: http.StatusOK})
	}
}

func (s *Server) getStartupConfig() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		cfg, err := config.GetStartupConfig()
		if err != nil {
			StatusCode := http.StatusInternalServerError
			if os.IsNotExist(err) {
				StatusCode = http.StatusNotFound
			}
			return s.error(w, StatusCode, fmt.Errorf("faiiled to get startup config: %v", err))
		}

		return s.respond(w, httpx.ConfigResponse{
			StatusCode: http.StatusOK,
			Config:     cfg,
		})
	}
}

func (s *Server) copyRunCfgToStartCfg() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		if err := config.CopyRunningConfigToStartupConfig(); err != nil {
			return s.error(w, http.StatusInternalServerError, fmt.Errorf("failed to copy running config to startup config: %v", err))
		}
		return s.respond(w, httpx.ConfigResponse{StatusCode: http.StatusOK, Config: config.GetRunningConfig()})
	}
}
