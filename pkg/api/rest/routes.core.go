package rest

import (
	"net/http"
	"time"
)

func (s *YambolHTTPServer) home() yambolHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) Response {
		resp := YambolStatsResponse{
			statusCode: http.StatusOK,
			Uptime:     time.Since(s.startedAt).String(),
			QueueStats: s.b.Stats(),
		}
		s.respond(w, resp)
		return resp
	}
}
