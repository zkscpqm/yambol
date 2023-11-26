package rest

import (
	"encoding/json"
	"net/http"
	"yambol/config"
	"yambol/pkg/transport/model"

	"yambol/pkg/telemetry"
)

func jMarshalIndent(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "    ")
}

type Response interface {
	GetStatusCode() int
	Render() ([]byte, error)
}

type ErrorResponse struct {
	StatusCode int
	Error      string `json:"error"`
}

func (r ErrorResponse) GetStatusCode() int {
	return r.StatusCode
}

func (r ErrorResponse) Render() ([]byte, error) {
	return jMarshalIndent(r)
}

type HomeResponse struct {
	StatusCode int
	model.BasicInfo
}

func (r HomeResponse) GetStatusCode() int {
	return r.StatusCode
}

func (r HomeResponse) Render() ([]byte, error) {
	return jMarshalIndent(r)
}

type ConfigResponse struct {
	StatusCode int
	Config     config.Configuration
}

func (r ConfigResponse) GetStatusCode() int {
	return r.StatusCode
}

func (r ConfigResponse) Render() ([]byte, error) {
	return jMarshalIndent(r.Config)
}

type StatsResponse map[string]telemetry.QueueStats

func (r StatsResponse) GetStatusCode() int {
	return http.StatusOK
}

func (r StatsResponse) Render() ([]byte, error) {
	return jMarshalIndent(r)
}

type QueuesGetResponse map[string]telemetry.QueueStats

func (r QueuesGetResponse) GetStatusCode() int {
	return 200
}

func (r QueuesGetResponse) Render() ([]byte, error) {
	return jMarshalIndent(r)
}

type QueueGetResponse struct {
	StatusCode int
	Data       string `json:"data"`
}

func (r QueueGetResponse) GetStatusCode() int {
	return r.StatusCode
}

func (r QueueGetResponse) Render() ([]byte, error) {
	return jMarshalIndent(r)
}

type EmptyResponse struct {
	StatusCode int
}

func (r EmptyResponse) GetStatusCode() int {
	return r.StatusCode
}

func (r EmptyResponse) Render() ([]byte, error) {
	return []byte{}, nil
}
