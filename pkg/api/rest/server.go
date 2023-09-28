package rest

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"yambol/pkg/broker"

	"github.com/gorilla/mux"
)

var forbiddenQueueNames = []string{
	"broadcast",
}

type yambolHandlerFunc = func(w http.ResponseWriter, r *http.Request) Response

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
	return &YambolHTTPServer{
		router:         mux.NewRouter(),
		b:              b,
		defaultHeaders: defaultHeaders,
		startedAt:      time.Now(),
	}
}

func (s *YambolHTTPServer) ServeHTTP(port int) error {
	s.routes()
	http.Handle("/", s.router)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func (s *YambolHTTPServer) routes() {
	s.route(
		"/",
		s.home(),
		debugPrintHook(),
	).Methods("GET")

	s.route(
		"/queues",
		s.queues(),
		debugPrintHook(),
	).Methods("GET", "POST")

	for _, qName := range s.b.Queues() {
		s.addQueueRoute(qName, debugPrintHook())
	}
}

func (s *YambolHTTPServer) hook(path string, wrapped yambolHandlerFunc, hooks ...Hook) http.HandlerFunc {
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

func (s *YambolHTTPServer) route(path string, handler yambolHandlerFunc, hooks ...Hook) *mux.Route {
	return s.router.HandleFunc(path, s.hook(path, handler, hooks...))
}

func (s *YambolHTTPServer) error(w http.ResponseWriter, status int, err error) Response {
	resp := ErrorResponse{status, err.Error()}
	s.respond(w, resp)
	return resp
}

func (s *YambolHTTPServer) respond(w http.ResponseWriter, response Response) {
	for k, v := range s.defaultHeaders {
		w.Header().Set(k, v)
	}
	w.Header().Set("Content-Type", "application/json")
	if b, err := response.JsonMarshalIndent(); err != nil {
		s.error(w, http.StatusInternalServerError, fmt.Errorf("failed to marshal response: %v", err))
	} else {
		w.WriteHeader(response.StatusCode())
		w.Write(b)
	}
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
