package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"yambol/config"
	"yambol/pkg/transport/httpx"
)

const (
	ClientUserAgent = "yambol-client"
	MaxTimeout      = time.Minute
)

type Client struct {
	Url            string
	client         *http.Client
	defaultTimeout time.Duration
}

func NewClient(url string, c *http.Client, defaultTimeout time.Duration) *Client {
	return &Client{
		strings.TrimSuffix(url, "/"),
		c,
		defaultTimeout,
	}
}

func (c *Client) context() (context.Context, context.CancelFunc) {
	to := c.defaultTimeout
	if to < 1 {
		to = MaxTimeout
	}
	return context.WithTimeout(context.Background(), to)
}

func (c *Client) headers() map[string]string {
	return map[string]string{
		"User-Agent":   ClientUserAgent,
		"Content-Type": "application/json; charset=utf-8",
		"Accept":       "application/json; charset=utf-8",
	}
}

func (c *Client) contentType() string {
	return fmt.Sprintf("application/json; charset=utf-8")
}

func (c *Client) ok(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func (c *Client) checkError(resp *http.Response) (err error) {
	var errResp *httpx.ErrorResponse
	if err = json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return nil
	}
	return fmt.Errorf(errResp.Error)
}

func (c *Client) get(ctx context.Context, url string, headers map[string]string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build GET request: %v", err)
	}
	return c.do(ctx, req, headers)
}

func (c *Client) post(ctx context.Context, url string, body io.Reader, headers map[string]string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to build POST request: %v", err)
	}
	return c.do(ctx, req, headers)
}

func (c *Client) put(ctx context.Context, url string, body io.Reader, headers map[string]string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to build PUT request: %v", err)
	}
	return c.do(ctx, req, headers)
}

func (c *Client) delete(ctx context.Context, url string, headers map[string]string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build DELETE request: %v", err)
	}
	return c.do(ctx, req, headers)
}

func (c *Client) do(ctx context.Context, req *http.Request, headers map[string]string) (resp *http.Response, err error) {
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	for key, value := range c.headers() {
		req.Header.Set(key, value)
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = c.context()
		defer cancel()
	}

	resp, err = c.client.Do(req.WithContext(ctx))
	if err != nil {
		return
	}
	if !c.ok(resp) {
		err = c.checkError(resp)
	}
	return
}

func (c *Client) Ping() (*httpx.HomeResponse, error) {
	ctx, cancel := c.context()
	defer cancel()
	return c.PingContext(ctx)
}

func (c *Client) PingContext(ctx context.Context) (*httpx.HomeResponse, error) {
	endpoint := httpx.UrlJoin(c.Url, "/")
	resp, err := c.get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping %s: %v", endpoint, err)
	}
	defer resp.Body.Close()
	if !c.ok(resp) {
		return nil, fmt.Errorf("[%d] failed to ping %s: %v", resp.StatusCode, endpoint, c.checkError(resp))
	}
	var response httpx.HomeResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ping response: %v", err)
	}
	return &response, nil
}

func (c *Client) Stats() (*httpx.StatsResponse, error) {
	ctx, cancel := c.context()
	defer cancel()
	return c.StatsContext(ctx)
}

func (c *Client) StatsContext(ctx context.Context) (*httpx.StatsResponse, error) {
	endpoint := httpx.UrlJoin(c.Url, "stats")
	resp, err := c.get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %v", err)
	}
	defer resp.Body.Close()
	if !c.ok(resp) {
		return nil, fmt.Errorf("[%d] failed to get stats: %v", resp.StatusCode, c.checkError(resp))
	}
	var response httpx.StatsResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode stats response: %v", err)
	}
	return &response, nil
}

func (c *Client) Publish(queue, value string) error {
	ctx, cancel := c.context()
	defer cancel()
	return c.PublishContext(ctx, queue, value)
}

func (c *Client) PublishContext(ctx context.Context, queue, value string) error {
	endpoint := httpx.UrlJoin(c.Url, "queues", queue)
	resp, err := c.post(ctx, endpoint, bytes.NewBufferString(value), nil)
	if err != nil {
		return fmt.Errorf("failed to send value to queue %s: %v", queue, err)
	}
	defer resp.Body.Close()
	if !c.ok(resp) {
		return fmt.Errorf("[%d] failed to send value to queue %s: %v", resp.StatusCode, queue, c.checkError(resp))
	}
	var response httpx.StatsResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return fmt.Errorf("failed to decode push response: %v", err)
	}
	return nil
}

func (c *Client) Consume(queue string) (string, error) {
	ctx, cancel := c.context()
	defer cancel()
	return c.ConsumeContext(ctx, queue)
}

func (c *Client) ConsumeContext(ctx context.Context, queue string) (string, error) {
	endpoint := httpx.UrlJoin(c.Url, "queues", queue)
	resp, err := c.get(ctx, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to consume from queue %s: %v", queue, err)
	}
	defer resp.Body.Close()
	if !c.ok(resp) {
		return "", fmt.Errorf("[%d] failed to consume value from queue %s: %v", resp.StatusCode, queue, c.checkError(resp))
	}
	var response httpx.QueueGetResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("failed to decode consume response: %v", err)
	}
	return response.Data, nil
}

func (c *Client) GetQueues() (*httpx.QueuesGetResponse, error) {
	ctx, cancel := c.context()
	defer cancel()
	return c.GetQueuesContext(ctx)
}

func (c *Client) GetQueuesContext(ctx context.Context) (*httpx.QueuesGetResponse, error) {

	endpoint := httpx.UrlJoin(c.Url, "queues")
	resp, err := c.get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get queues: %v", err)
	}
	defer resp.Body.Close()
	if !c.ok(resp) {
		return nil, fmt.Errorf("[%d] failed to get queues: %v", resp.StatusCode, c.checkError(resp))
	}
	var queues httpx.QueuesGetResponse
	err = json.NewDecoder(resp.Body).Decode(&queues)
	if err != nil {
		return nil, fmt.Errorf("failed to decode queues response: %v", err)
	}
	return &queues, nil
}

func (c *Client) CreateQueueContext(ctx context.Context, queue string, opts config.QueueConfig) error {
	qInfo := httpx.QueuesPostRequest{
		Name:        queue,
		QueueConfig: opts,
	}
	endpoint := httpx.UrlJoin(c.Url, "queues")
	b, err := json.Marshal(qInfo)
	if err != nil {
		return fmt.Errorf("failed to serialize queue info: %v", err)
	}
	resp, err := c.post(ctx, endpoint, bytes.NewBuffer(b), nil)
	if err != nil {
		return fmt.Errorf("failed to create queue %s: %v", queue, err)
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("[%d] failed to create queue %s: %v", resp.StatusCode, queue, c.checkError(resp))
	}
	return nil
}

func (c *Client) DeleteQueue(queue string) error {
	ctx, cancel := c.context()
	defer cancel()
	return c.DeleteQueueContext(ctx, queue)
}

func (c *Client) DeleteQueueContext(ctx context.Context, queue string) error {

	endpoint := httpx.UrlJoin(c.Url, "queues", queue)
	resp, err := c.delete(ctx, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to delete queue %s: %v", queue, err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[%d] failed to delete queue %s: %v", resp.StatusCode, queue, c.checkError(resp))
	}
	return nil
}

func (c *Client) CreateQueue(queue string, opts config.QueueConfig) error {
	ctx, cancel := c.context()
	defer cancel()
	return c.CreateQueueContext(ctx, queue, opts)
}

func (c *Client) GetRunningConfig() (*config.Configuration, error) {

	ctx, cancel := c.context()
	defer cancel()
	return c.GetRunningConfigContext(ctx)
}

func (c *Client) GetRunningConfigContext(ctx context.Context) (*config.Configuration, error) {
	endpoint := httpx.UrlJoin(c.Url, "running_config")
	resp, err := c.get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get running config: %v", err)
	}
	defer resp.Body.Close()
	if !c.ok(resp) {
		return nil, fmt.Errorf("[%d] failed to get running config: %v", resp.StatusCode, c.checkError(resp))
	}
	var cfg config.Configuration
	err = json.NewDecoder(resp.Body).Decode(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode running config response: %v", err)
	}
	return &cfg, nil
}

func (c *Client) SetRunningConfig(cfg config.Configuration) (err error) {
	endpoint := httpx.UrlJoin(c.Url, "running_config")

	var buffer bytes.Buffer
	if err = json.NewEncoder(&buffer).Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config: %v", err)
	}

	resp, err := c.client.Post(endpoint, c.contentType(), &buffer)
	if err != nil {
		return fmt.Errorf("failed to POST running config: %v", err)
	}
	defer resp.Body.Close()
	if !c.ok(resp) {
		return fmt.Errorf("[%d] failed to get running config: %s", resp.StatusCode, c.checkError(resp))
	}
	return nil
}

func (c *Client) SetRunningConfigContext(ctx context.Context, cfg config.Configuration) (err error) {
	endpoint := httpx.UrlJoin(c.Url, "running_config")

	var buffer bytes.Buffer
	if err = json.NewEncoder(&buffer).Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config: %v", err)
	}

	resp, err := c.post(ctx, endpoint, &buffer, nil)
	if err != nil {
		return fmt.Errorf("failed to POST running config: %v", err)
	}
	defer resp.Body.Close()
	if !c.ok(resp) {
		return fmt.Errorf("[%d] failed to get running config: %s", resp.StatusCode, c.checkError(resp))
	}
	return nil
}

func (c *Client) GetStartupConfig() (*config.Configuration, error) {

	ctx, cancel := c.context()
	defer cancel()
	return c.GetStartupConfigContext(ctx)
}

func (c *Client) GetStartupConfigContext(ctx context.Context) (*config.Configuration, error) {
	endpoint := httpx.UrlJoin(c.Url, "startup_config")
	resp, err := c.get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get startup config: %v", err)
	}
	defer resp.Body.Close()
	if !c.ok(resp) {
		return nil, fmt.Errorf("[%d] failed to get startup config: %v", resp.StatusCode, c.checkError(resp))
	}
	var cfg config.Configuration
	err = json.NewDecoder(resp.Body).Decode(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode startup config response: %v", err)
	}
	return &cfg, nil
}

func (c *Client) CopyRunCfgToStartCfg() (cfg *config.Configuration, err error) {
	ctx, cancel := c.context()
	defer cancel()
	return c.CopyRunCfgToStartCfgContext(ctx)
}

func (c *Client) CopyRunCfgToStartCfgContext(ctx context.Context) (*config.Configuration, error) {
	endpoint := httpx.UrlJoin(c.Url, "running_config", "save")

	resp, err := c.put(ctx, endpoint, bytes.NewReader([]byte{}), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to update startup config: %v", err)
	}
	defer resp.Body.Close()
	if !c.ok(resp) {
		return nil, fmt.Errorf("[%d] failed to update startup config: %v", resp.StatusCode, c.checkError(resp))
	}
	var cfg config.Configuration
	err = json.NewDecoder(resp.Body).Decode(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode new startup config response: %v", err)
	}
	return &cfg, nil
}
