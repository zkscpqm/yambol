package fe

import (
	"fmt"
	"github.com/CloudyKit/jet"
	"net/http"
	"path"
	"time"
	"yambol/pkg/util"
	"yambol/pkg/util/log"

	"yambol/pkg/broker"
	"yambol/pkg/transport/httpx"

	"github.com/gorilla/mux"
)

var (
	views = jet.NewHTMLSet(
		path.Join(util.ProjectRootDirectoryMust(), "views"),
	)
)

func init() {
	views.SetDevelopmentMode(true)
}

type Server struct {
	router         *mux.Router
	b              *broker.MessageBroker
	defaultHeaders map[string]string
	startedAt      time.Time
	logger         *log.Logger
}

func NewServer(b *broker.MessageBroker, defaultHeaders map[string]string, logger *log.Logger) *Server {
	if defaultHeaders == nil {
		defaultHeaders = make(map[string]string)
	}
	rtr := mux.NewRouter()

	return &Server{
		router:         rtr,
		b:              b,
		defaultHeaders: defaultHeaders,
		logger:         logger.NewFrom("HTTP"),
	}
}

func (s *Server) ListenAndServeInsecure(port int) error {
	return s.ListenAndServe(port, "", "")
}

func (s *Server) ListenAndServe(port int, certFile, keyFile string) error {
	s.logger.Info("trying to listen on [%d]...", port)
	s.routes()
	addr := fmt.Sprintf(":%d", port)
	s.startedAt = time.Now()
	if certFile == "" || keyFile == "" {
		s.logger.Info("Starting Yambol Frontend with http (insecure) at [%d]", port)
		return http.ListenAndServe(addr, s.router)
	}
	s.logger.Info("Starting Yambol Frontend with https (secure) at [%d]", port)
	return http.ListenAndServeTLS(addr, certFile, keyFile, s.router)
}

func (s *Server) routes() {
	s.hostLocalStatic()
	s.route(
		"/",
		s.home(),
		httpx.DebugPrintHook(s.logger.WithLevel(log.LevelInfo)), // To avoid printing huge HTML return values
	).Methods(http.MethodGet)
	s.route(
		"/queues",
		s.queues(),
		httpx.DebugPrintHook(s.logger.WithLevel(log.LevelInfo)),
	).Methods(http.MethodGet)
}

func (s *Server) hook(path string, wrapped httpx.HandlerFunc, hooks ...httpx.Middleware) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		for _, h := range hooks {
			ok, err := h.Before(req)

			if err != nil {
				s.logger.Error("error in %s pre-execution hook: %v", path, err)
			}
			if !ok {
				return
			}
		}
		resp := wrapped(w, req)
		for _, h := range hooks {
			if err := h.After(req, resp); err != nil {
				s.logger.Error("error in %s post-execution hook: %v", path, err)
			}
		}
	}
}

func (s *Server) route(path string, handler httpx.HandlerFunc, hooks ...httpx.Middleware) *mux.Route {
	s.logger.Debug("routing [%s]", path)
	return s.router.HandleFunc(path, s.hook(path, handler, hooks...))
}

func (s *Server) error(w http.ResponseWriter, status int, err error, args ...any) Response {
	return s.respond(w, ErrorResponse{StatusCode: status, Error: err.Error() + fmt.Sprint(args...)})
}

func (s *Server) respond(w http.ResponseWriter, response Response) Response {
	for k, v := range s.defaultHeaders {
		w.Header().Set(k, v)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if b, err := response.Render(); err != nil {
		if errResp, ok := response.(ErrorResponse); ok {
			// worst case scenario. Yes it actually happened
			w.WriteHeader(errResp.GetStatusCode())
			w.Write([]byte(errResp.Error))
			return errResp
		}
		return s.error(w, http.StatusInternalServerError, fmt.Errorf("failed to marshal response: %v", err))
	} else {
		w.WriteHeader(response.GetStatusCode())
		w.Write(b)
	}
	return response
}

func (s *Server) isError(r Response) bool {
	return r.GetStatusCode() > 399
}
