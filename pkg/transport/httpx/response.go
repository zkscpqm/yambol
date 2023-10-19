package httpx

import (
	"encoding/json"
	"net/http"
	"yambol/config"

	"yambol/pkg/telemetry"
)

func jMarshalIndent(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "    ")
}

type Response interface {
	GetStatusCode() int
	JsonMarshalIndent() ([]byte, error)
}

type ErrorResponse struct {
	StatusCode int
	Error      string `json:"error"`
}

func (r ErrorResponse) GetStatusCode() int {
	return r.StatusCode
}

func (r ErrorResponse) JsonMarshalIndent() ([]byte, error) {
	return jMarshalIndent(r)
}

type YambolHomeResponse struct {
	StatusCode int
	Uptime     string `json:"uptime"`
	Version    string `json:"version"`
}

func (r YambolHomeResponse) GetStatusCode() int {
	return r.StatusCode
}

func (r YambolHomeResponse) JsonMarshalIndent() ([]byte, error) {
	return jMarshalIndent(r)
}

type YambolConfigResponse struct {
	StatusCode int
	Config     config.Configuration
}

func (r YambolConfigResponse) GetStatusCode() int {
	return r.StatusCode
}

func (r YambolConfigResponse) JsonMarshalIndent() ([]byte, error) {
	return jMarshalIndent(r.Config)
}

type YambolStatsResponse map[string]telemetry.QueueStats

func (r YambolStatsResponse) GetStatusCode() int {
	return http.StatusOK
}

func (r YambolStatsResponse) JsonMarshalIndent() ([]byte, error) {
	return jMarshalIndent(r)
}

type QueuesGetResponse map[string]telemetry.QueueStats

func (r QueuesGetResponse) GetStatusCode() int {
	return 200
}

func (r QueuesGetResponse) JsonMarshalIndent() ([]byte, error) {
	return jMarshalIndent(r)
}

type QueueGetResponse struct {
	StatusCode int
	Data       string `json:"data"`
}

func (r QueueGetResponse) GetStatusCode() int {
	return r.StatusCode
}

func (r QueueGetResponse) JsonMarshalIndent() ([]byte, error) {
	return jMarshalIndent(r)
}

type EmptyResponse struct {
	StatusCode int
}

func (r EmptyResponse) GetStatusCode() int {
	return r.StatusCode
}

func (r EmptyResponse) JsonMarshalIndent() ([]byte, error) {
	return []byte{}, nil
}
