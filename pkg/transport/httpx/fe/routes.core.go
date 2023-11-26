package fe

import (
	"net/http"
	"path"
	"time"
	"yambol/pkg/transport/httpx"
	"yambol/pkg/transport/model"

	"yambol/pkg/util"
)

func (s *Server) hostLocalStatic() {
	s.router.Handle(
		"/static/{_:.*}",
		http.StripPrefix(
			"/static/",
			http.FileServer(
				http.Dir(
					path.Join(
						util.ProjectRootDirectoryMust(), "static",
					),
				),
			),
		),
	)
}

func (s *Server) home() httpx.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		return s.respond(w, HomeResponse{
			StatusCode: http.StatusOK,
			BasicInfo: model.BasicInfo{
				Uptime:  time.Since(s.startedAt),
				Version: util.Version(),
			},
		})
	}
}

func (s *Server) queues() httpx.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		return s.respond(w, StatsResponse{
			stats: s.b.Stats(),
		})
	}
}
