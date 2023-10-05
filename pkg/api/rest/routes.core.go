package rest

import (
	"net/http"
	"time"
	"yambol/pkg/util"
)

func (s *YambolHTTPServer) home() yambolHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) Response {
		resp := YambolHomeResponse{
			statusCode: http.StatusOK,
			Uptime:     time.Since(s.startedAt).String(),
			Version:    util.Version(),
		}
		s.respond(w, resp)
		return resp
	}
}

func (s *YambolHTTPServer) stats() yambolHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) Response {
		resp := YambolStatsResponse(s.b.Stats())

		s.respond(w, resp)
		return resp
	}
}
