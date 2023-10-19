package rest

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"yambol/pkg/transport/httpx"

	"yambol/pkg/broker"

	"github.com/gorilla/mux"
)

var forbiddenQueueNames = []string{
	"broadcast",
}

type yambolHandlerFunc = func(w http.ResponseWriter, r *http.Request) httpx.Response

type YambolHTTPServer struct {
	router         *mux.Router
	b              *broker.MessageBroker
	defaultHeaders map[string]string
	startedAt      time.Time
}

func NewYambolHTTPServer(b *broker.MessageBroker, defaultHeaders map[string]string) *YambolHTTPServer {
	if defaultHeaders == nil {
		defaultHeaders = make(map[string]string)
	}
	rtr := mux.NewRouter()
	rtr.Use(httpx.LoggingMiddleware)
	return &YambolHTTPServer{
		router:         rtr,
		b:              b,
		defaultHeaders: defaultHeaders,
	}
}

func (s *YambolHTTPServer) ListenAndServeInsecure(port int) error {
	return s.ListenAndServe(port, "", "")
}

func (s *YambolHTTPServer) ListenAndServe(port int, certFile, keyFile string) error {
	s.routes()
	addr := fmt.Sprintf(":%d", port)
	s.startedAt = time.Now()
	if certFile == "" || keyFile == "" {
		fmt.Printf("Starting Yambol with http (insecure) at [%d]\n", port)

		return http.ListenAndServe(addr, s.router)
	}
	fmt.Printf("Starting Yambol with https (secure) at [%d]\n", port)
	return http.ListenAndServeTLS(addr, certFile, keyFile, s.router)
}

func (s *YambolHTTPServer) routes() {
	s.route(
		"/",
		s.home(),
		httpx.DebugPrintHook(),
	).Methods("GET")

	s.route(
		"/stats",
		s.stats(),
		httpx.DebugPrintHook(),
	).Methods("GET")

	s.route(
		"/queues",
		s.queues(),
		httpx.DebugPrintHook(),
	).Methods("GET", "POST")

	s.route(
		"/running_config",
		s.runningConfig(),
		httpx.DebugPrintHook(),
	).Methods("GET", "POST")

	s.route(
		"/startup_config",
		s.getStartupConfig(),
		httpx.DebugPrintHook(),
	).Methods("GET")

	s.route(
		"/running_config/save",
		s.copyRunCfgToStartCfg(),
		httpx.DebugPrintHook(),
	).Methods("PUT")

	for _, qName := range s.b.Queues() {
		s.addQueueRoute(qName, httpx.DebugPrintHook())
	}
}

func (s *YambolHTTPServer) hook(path string, wrapped yambolHandlerFunc, hooks ...httpx.Hook) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		for _, h := range hooks {
			ok, err := h.Before(req)

			if err != nil {
				fmt.Printf("error in %s pre-execution hook: %v\n", path, err)
			}
			if !ok {
				return
			}
		}
		resp := wrapped(w, req)
		for _, h := range hooks {
			if err := h.After(req, resp); err != nil {
				fmt.Printf("error in %s post-execution hook: %v\n", path, err)
			}
		}
	}
}

func (s *YambolHTTPServer) route(path string, handler yambolHandlerFunc, hooks ...httpx.Hook) *mux.Route {
	return s.router.HandleFunc(path, s.hook(path, handler, hooks...))
}

func (s *YambolHTTPServer) error(w http.ResponseWriter, status int, err error, args ...any) httpx.Response {
	resp := httpx.ErrorResponse{status, err.Error() + fmt.Sprint(args...)}
	s.respond(w, resp)
	return resp
}

func (s *YambolHTTPServer) respond(w http.ResponseWriter, response httpx.Response) httpx.Response {
	for k, v := range s.defaultHeaders {
		w.Header().Set(k, v)
	}
	w.Header().Set("Content-Type", "application/json")
	if b, err := response.JsonMarshalIndent(); err != nil {
		s.error(w, http.StatusInternalServerError, fmt.Errorf("failed to marshal response: %v", err))
	} else {
		w.WriteHeader(response.GetStatusCode())
		w.Write(b)
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
