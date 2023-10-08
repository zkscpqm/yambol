package rest

import (
	"encoding/json"
	"net/http"

	"yambol/pkg/telemetry"
)

type Response interface {
	StatusCode() int
	JsonMarshalIndent() ([]byte, error)
}

type ErrorResponse struct {
	statusCode int
	Error      string `json:"error"`
}

func (r ErrorResponse) StatusCode() int {
	return r.statusCode
}

func (r ErrorResponse) JsonMarshalIndent() ([]byte, error) {
	return json.MarshalIndent(r, "", "    ")
}

type YambolHomeResponse struct {
	statusCode int
	Uptime     string `json:"uptime"`
	Version    string `json:"version"`
}

func (r YambolHomeResponse) StatusCode() int {
	return r.statusCode
}

func (r YambolHomeResponse) JsonMarshalIndent() ([]byte, error) {
	return json.MarshalIndent(r, "", "    ")
}

type YambolStatsResponse map[string]telemetry.QueueStats

func (r YambolStatsResponse) StatusCode() int {
	return http.StatusOK
}

func (r YambolStatsResponse) JsonMarshalIndent() ([]byte, error) {
	return json.MarshalIndent(r, "", "    ")
}

type QueuesGetResponse map[string]telemetry.QueueStats

func (r QueuesGetResponse) StatusCode() int {
	return 200
}

func (r QueuesGetResponse) JsonMarshalIndent() ([]byte, error) {
	return json.MarshalIndent(r, "", "    ")
}

type QueueGetResponse struct {
	statusCode int
	Data       string `json:"data"`
}

func (r QueueGetResponse) StatusCode() int {
	return r.statusCode
}

func (r QueueGetResponse) JsonMarshalIndent() ([]byte, error) {
	return json.MarshalIndent(r, "", "    ")
}

type EmptyResponse struct {
	statusCode int
}

func (r EmptyResponse) StatusCode() int {
	return r.statusCode
}

func (r EmptyResponse) JsonMarshalIndent() ([]byte, error) {
	return []byte{}, nil
}
