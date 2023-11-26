package fe

import (
	"bytes"
	"fmt"
	"github.com/CloudyKit/jet"
	"net/http"
	"yambol/pkg/telemetry"
	"yambol/pkg/transport/httpx"
	"yambol/pkg/transport/model"
)

func jetRender(t *jet.Template) ([]byte, jet.VarMap, error) {
	var buf bytes.Buffer
	vars := make(jet.VarMap)
	if err := t.Execute(&buf, vars, nil); err != nil {
		return nil, nil, fmt.Errorf("failed to render JET template `%s`: %v", t.Name, err)
	}
	return buf.Bytes(), vars, nil
}

type Response interface {
	Template() (*jet.Template, error)
	Render() (b []byte, err error)
	httpx.Response
}

type ErrorResponse struct {
	StatusCode int
	vars       jet.VarMap
	Error      string `json:"error"`
}

func (r ErrorResponse) Template() (*jet.Template, error) {
	return views.GetTemplate("error.jet")
}

func (r ErrorResponse) GetStatusCode() int {
	return r.StatusCode
}

func (r ErrorResponse) Render() (b []byte, err error) {
	t, err := r.Template()
	if err != nil {
		return nil, fmt.Errorf("failed to get ERROR template: %v", err)
	}
	b, r.vars, err = jetRender(t)
	return
}

type HomeResponse struct {
	StatusCode int
	vars       jet.VarMap
	model.BasicInfo
}

func (r HomeResponse) Template() (*jet.Template, error) {
	return views.GetTemplate("home.jet")
}

func (r HomeResponse) GetStatusCode() int {
	return r.StatusCode
}

func (r HomeResponse) Render() (b []byte, err error) {
	t, err := r.Template()
	if err != nil {
		return nil, fmt.Errorf("failed to get HOME template: %v", err)
	}
	b, r.vars, err = jetRender(t)
	return
}

type StatsResponse struct {
	stats map[string]telemetry.QueueStats
	vars  jet.VarMap
}

func (r StatsResponse) Template() (*jet.Template, error) {
	return views.GetTemplate("queues.jet")
}

func (r StatsResponse) GetStatusCode() int {
	return http.StatusOK
}

func (r StatsResponse) Render() (b []byte, err error) {
	t, err := r.Template()
	if err != nil {
		return nil, fmt.Errorf("failed to get QUEUES template: %v", err)
	}
	b, r.vars, err = jetRender(t)
	return
}
