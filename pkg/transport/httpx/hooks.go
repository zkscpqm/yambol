package httpx

import (
	"fmt"
	"net/http"
)

type RequestHook func(*http.Request) (bool, error)
type ResponseHook func(*http.Request, Response) error

type Hook struct {
	name   string
	Before RequestHook
	After  ResponseHook
}

func NewHook(name string, before RequestHook, after ResponseHook) Hook {
	if before == nil {
		before = func(*http.Request) (bool, error) { return true, nil }
	}
	if after == nil {
		after = func(*http.Request, Response) error { return nil }
	}
	return Hook{
		name:   name,
		Before: before,
		After:  after,
	}
}

func DebugPrintHook() Hook {
	return NewHook("debug_print",
		func(req *http.Request) (bool, error) {
			fmt.Printf("%s -> %s %s\n", req.RemoteAddr, req.Method, req.URL.Path)
			return true, nil
		},
		func(req *http.Request, resp Response) error {
			fmt.Printf("%s <- %s %s [%d]\n", req.RemoteAddr, req.Method, req.URL.Path, resp.GetStatusCode())
			return nil
		},
	)
}
