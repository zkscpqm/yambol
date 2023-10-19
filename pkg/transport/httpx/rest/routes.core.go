package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"yambol/config"
	"yambol/pkg/transport/httpx"

	"yambol/pkg/util"
)

func (s *YambolHTTPServer) home() yambolHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		return s.respond(w, httpx.YambolHomeResponse{
			StatusCode: http.StatusOK,
			Uptime:     time.Since(s.startedAt).String(),
			Version:    util.Version(),
		})
	}
}

func (s *YambolHTTPServer) stats() yambolHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		return s.respond(w, httpx.YambolStatsResponse(s.b.Stats()))
	}
}

func (s *YambolHTTPServer) runningConfig() yambolHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		target, err := resolveHTTPMethodTarget(r, map[string]yambolHandlerFunc{
			"GET":  s.getRunningConfig(),
			"POST": s.setRunningConfig(),
		})
		if err != nil {
			return s.error(w, http.StatusMethodNotAllowed, err, r.Method)
		}
		return target(w, r)
	}
}

func (s *YambolHTTPServer) getRunningConfig() yambolHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		return s.respond(w, httpx.YambolConfigResponse{
			StatusCode: http.StatusOK,
			Config:     config.GetRunningConfig(),
		})
	}
}

func (s *YambolHTTPServer) setRunningConfig() yambolHandlerFunc {
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

func (s *YambolHTTPServer) startupConfig() yambolHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		target, err := resolveHTTPMethodTarget(r, map[string]yambolHandlerFunc{
			"GET":  s.getStartupConfig(),
			"POST": s.setRunningConfig(),
		})
		if err != nil {
			return s.error(w, http.StatusMethodNotAllowed, err, r.Method)
		}
		return target(w, r)
	}
}

func (s *YambolHTTPServer) getStartupConfig() yambolHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		cfg, err := config.GetStartupConfig()
		if err != nil {
			StatusCode := http.StatusInternalServerError
			if os.IsNotExist(err) {
				StatusCode = http.StatusNotFound
			}
			return s.error(w, StatusCode, fmt.Errorf("faiiled to get startup config: %v", err))
		}

		return s.respond(w, httpx.YambolConfigResponse{
			StatusCode: http.StatusOK,
			Config:     cfg,
		})
	}
}

func (s *YambolHTTPServer) copyRunCfgToStartCfg() yambolHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		if err := config.CopyRunningConfigToStartupConfig(); err != nil {
			return s.error(w, http.StatusInternalServerError, fmt.Errorf("failed to copy running config to startup config: %v", err))
		}
		return s.respond(w, httpx.YambolConfigResponse{StatusCode: http.StatusOK, Config: config.GetRunningConfig()})
	}
}
