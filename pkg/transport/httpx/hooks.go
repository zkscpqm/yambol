package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
			logger.Info("%s -> %s %s", req.RemoteAddr, req.Method, req.URL.Path)
			if logger.GetLevel() <= log.LevelDebug {
				reqBody, err := safeJsonRequestBodyReader(req)
				if err != nil {
					return false, fmt.Errorf("failed to read request body: %v", err)
				}
				if reqBody != "" {
					logger.Debug("Request Body:\n%s", reqBody)
				}
			}
			return true, nil
		},
		func(req *http.Request, resp Response) (err error) {

			logger.Info("%s <- %s %s [%d]", req.RemoteAddr, req.Method, req.URL.Path, resp.GetStatusCode())

			// Avoid doing this JSON marshaling if we are not in debug mode to begin with
			if logger.GetLevel() <= log.LevelDebug {
				responseBody := make([]byte, 0)
				responseBody, err = resp.Render()
				logger.Debug("Response Body:\n%s", string(responseBody))
			}

			return
		},
	)
}

func safeJsonRequestBodyReader(req *http.Request) (string, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read request body")
	}
	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewBuffer(body))

	if string(body) == "" {
		return "", nil
	}
	var data map[string]interface{}
	if err = json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON body: %v", err)
	}

	// MarshalIndent to get formatted JSON
	formattedJSON, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return "", fmt.Errorf("failed to re-marshal JSON to string")
	}
	return string(formattedJSON), nil
}
