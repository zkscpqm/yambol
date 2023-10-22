package rest

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"yambol/pkg/util/log"

	"yambol/pkg/broker"
	"yambol/pkg/transport/httpx"

	"github.com/gorilla/mux"
)

var forbiddenQueueNames = []string{
	"broadcast",
}

type yambolHandlerFunc = func(w http.ResponseWriter, r *http.Request) httpx.Response

type YambolRESTServer struct {
	router         *mux.Router
	b              *broker.MessageBroker
	defaultHeaders map[string]string
	startedAt      time.Time
	logger         *log.Logger
}

func NewYambolRESTServer(b *broker.MessageBroker, defaultHeaders map[string]string, logger *log.Logger) *YambolRESTServer {
	if defaultHeaders == nil {
		defaultHeaders = make(map[string]string)
	}
	rtr := mux.NewRouter()

	return &YambolRESTServer{
		router:         rtr,
		b:              b,
		defaultHeaders: defaultHeaders,
		logger:         logger.NewFrom("REST"),
	}
}

func (s *YambolRESTServer) ListenAndServeInsecure(port int) error {
	return s.ListenAndServe(port, "", "")
}

func (s *YambolRESTServer) ListenAndServe(port int, certFile, keyFile string) error {
	s.logger.Info("trying to listen on [%d]...", port)
	s.routes()
	addr := fmt.Sprintf(":%d", port)
	s.startedAt = time.Now()
	if certFile == "" || keyFile == "" {
		s.logger.Info("Starting Yambol with http (insecure) at [%d]", port)
		return http.ListenAndServe(addr, s.router)
	}
	s.logger.Info("Starting Yambol with https (secure) at [%d]", port)
	return http.ListenAndServeTLS(addr, certFile, keyFile, s.router)
}

func (s *YambolRESTServer) routes() {
	s.route(
		"/",
		s.home(),
		httpx.DebugPrintHook(s.logger),
	).Methods("GET")

	s.route(
		"/stats",
		s.stats(),
		httpx.DebugPrintHook(s.logger),
	).Methods("GET")

	s.route(
		"/queues",
		s.queues(),
		httpx.DebugPrintHook(s.logger),
	).Methods("GET", "POST")

	s.route(
		"/running_config",
		s.runningConfig(),
		httpx.DebugPrintHook(s.logger),
	).Methods("GET", "POST")

	s.route(
		"/startup_config",
		s.getStartupConfig(),
		httpx.DebugPrintHook(s.logger),
	).Methods("GET")

	s.route(
		"/running_config/save",
		s.copyRunCfgToStartCfg(),
		httpx.DebugPrintHook(s.logger),
	).Methods("PUT")

	for _, qName := range s.b.Queues() {
		s.addQueueRoute(qName, httpx.DebugPrintHook(s.logger))
	}
}

func (s *YambolRESTServer) hook(path string, wrapped yambolHandlerFunc, hooks ...httpx.Middleware) http.HandlerFunc {
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

func (s *YambolRESTServer) route(path string, handler yambolHandlerFunc, hooks ...httpx.Middleware) *mux.Route {
	s.logger.Debug("routing [%s]", path)
	return s.router.HandleFunc(path, s.hook(path, handler, hooks...))
}

func (s *YambolRESTServer) error(w http.ResponseWriter, status int, err error, args ...any) httpx.Response {
	return s.respond(w, httpx.ErrorResponse{StatusCode: status, Error: err.Error() + fmt.Sprint(args...)})
}

func (s *YambolRESTServer) respond(w http.ResponseWriter, response httpx.Response) httpx.Response {
	for k, v := range s.defaultHeaders {
		w.Header().Set(k, v)
	}
	w.Header().Set("Content-Type", "application/json")
	if b, err := response.JsonMarshalIndent(); err != nil {
		return s.error(w, http.StatusInternalServerError, fmt.Errorf("failed to marshal response: %v", err))
	} else {
		w.WriteHeader(response.GetStatusCode())
		w.Write(b)
		s.logger.Debug("wrote to buffer:\n[%d]\n%s", response.GetStatusCode(), string(b))
	}
	return response
}

func resolveHTTPMethodTarget(r *http.Request, targets map[string]yambolHandlerFunc) (yambolHandlerFunc, error) {
	allowedMethods := make([]string, 0, len(targets))
	for k := range targets {
		allowedMethods = append(allowedMethods, k)
	}
	target, ok := targets[r.Method]
	if !ok {
		return nil, fmt.Errorf("method %s not allowed on (%s), allowed methods: %v", r.Method, r.URL.Path, allowedMethods)
	}
	return target, nil

}

func normalizeQueueName(name string) string {
	return strings.ToLower(
		strings.TrimPrefix(
			strings.TrimSuffix(
				name,
				"/",
			),
			"/",
		),
	)
}

func isValidPath(name string) bool {
	name = normalizeQueueName(name)
	if strings.TrimSpace(name) == "" {
		return false
	}

	for _, forbiddenName := range forbiddenQueueNames {
		if name == forbiddenName {
			return false
		}
	}

	re := regexp.MustCompile(`^[\w-]+$`)

	return re.MatchString(name)
}
