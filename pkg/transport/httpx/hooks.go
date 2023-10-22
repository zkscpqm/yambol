package httpx

import (
	"net/http"

	"yambol/pkg/util/log"
)

type RequestHook func(*http.Request) (bool, error)
type ResponseHook func(*http.Request, Response) error

type Middleware struct {
	name   string
	Before RequestHook
	After  ResponseHook
}

func NewHook(name string, before RequestHook, after ResponseHook) Middleware {
	if before == nil {
		before = func(*http.Request) (bool, error) { return true, nil }
	}
	if after == nil {
		after = func(*http.Request, Response) error { return nil }
	}
	return Middleware{
		name:   name,
		Before: before,
		After:  after,
	}
}

// ---- implementations

func DebugPrintHook(logger *log.Logger) Middleware {
	return NewHook("debug_print",
		func(req *http.Request) (bool, error) {
			logger.Debug("%s -> %s %s\n", req.RemoteAddr, req.Method, req.URL.Path)
			return true, nil
		},
		func(req *http.Request, resp Response) error {
			logger.Debug("%s <- %s %s [%d]\n", req.RemoteAddr, req.Method, req.URL.Path, resp.GetStatusCode())
			return nil
		},
	)
}
