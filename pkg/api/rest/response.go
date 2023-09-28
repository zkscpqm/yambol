package rest

import (
	"encoding/json"
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

type YambolStatsResponse struct {
	statusCode int
	Uptime     string                          `json:"uptime"`
	QueueStats map[string]telemetry.QueueStats `json:"queue_stats"`
}

func (r YambolStatsResponse) StatusCode() int {
	return r.statusCode
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
