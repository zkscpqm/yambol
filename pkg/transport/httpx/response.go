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

type HomeResponse struct {
	StatusCode int
	Uptime     string `json:"uptime"`
	Version    string `json:"version"`
}

func (r HomeResponse) GetStatusCode() int {
	return r.StatusCode
}

func (r HomeResponse) JsonMarshalIndent() ([]byte, error) {
	return jMarshalIndent(r)
}

type ConfigResponse struct {
	StatusCode int
	Config     config.Configuration
}

func (r ConfigResponse) GetStatusCode() int {
	return r.StatusCode
}

func (r ConfigResponse) JsonMarshalIndent() ([]byte, error) {
	return jMarshalIndent(r.Config)
}

type StatsResponse map[string]telemetry.QueueStats

func (r StatsResponse) GetStatusCode() int {
	return http.StatusOK
}

func (r StatsResponse) JsonMarshalIndent() ([]byte, error) {
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
	return jMarshalIndent(r.Data)
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
