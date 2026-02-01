package http_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gojek/heimdall/v7"
	"go.uber.org/zap"

	"berlin.allobank.local/common/gommon/logger"
	"berlin.allobank.local/common/gommon/masking"
)

type RestClient interface {
	Get(ctx context.Context, path string, payload *RequestPayload) (*http.Response, error)
	Post(ctx context.Context, path string, payload *RequestPayload) (*http.Response, error)
	Put(ctx context.Context, path string, payload *RequestPayload) (*http.Response, error)
	Patch(ctx context.Context, path string, payload *RequestPayload) (*http.Response, error)
	Delete(ctx context.Context, path string, payload *RequestPayload) (*http.Response, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	AddPlugin(p heimdall.Plugin)
	ParseResponseBody(resp *http.Response, targetStruct any) error
}

type restClient struct {
	baseURL        string
	clientID       string
	defaultHeaders http.Header
	client         heimdall.Client
	logger         logger.PublicLoggerFn
}

func NewRestClient(clientID string, baseURL string, heimdalClient heimdall.Client, fnLogger logger.PublicLoggerFn) RestClient {
	if fnLogger == nil {
		fnLogger = func(ctx context.Context, identifier string, objects ...zap.Field) {
			log.Println(identifier)
		}
	}

	return &restClient{
		baseURL:  baseURL,
		clientID: clientID,
		client:   heimdalClient,
		logger:   fnLogger,
	}
}

func (c *restClient) Get(ctx context.Context, path string, payload *RequestPayload) (*http.Response, error) {
	request, err := c.prepareRequest(ctx, http.MethodGet, path, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	return c.Do(ctx, request)
}

func (c *restClient) Post(ctx context.Context, path string, payload *RequestPayload) (*http.Response, error) {
	request, err := c.prepareRequest(ctx, http.MethodPost, path, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	return c.Do(ctx, request)
}

func (c *restClient) Put(ctx context.Context, path string, payload *RequestPayload) (*http.Response, error) {
	request, err := c.prepareRequest(ctx, http.MethodPut, path, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	return c.Do(ctx, request)
}

func (c *restClient) Patch(ctx context.Context, path string, payload *RequestPayload) (*http.Response, error) {
	request, err := c.prepareRequest(ctx, http.MethodPatch, path, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	return c.Do(ctx, request)
}

func (c *restClient) Delete(ctx context.Context, path string, payload *RequestPayload) (*http.Response, error) {
	request, err := c.prepareRequest(ctx, http.MethodDelete, path, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	return c.Do(ctx, request)
}

func (c *restClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	startTime := time.Now()
	c.BeforeRequest(ctx, req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Do request failed: %w", err)
	}

	c.AfterResponse(ctx, startTime, req, resp)

	return resp, nil
}

func (c *restClient) AddPlugin(p heimdall.Plugin) {
	c.client.AddPlugin(p)
}

func (c *restClient) BeforeRequest(ctx context.Context, req *http.Request) {
	if c.logger == nil {
		return
	}

	var (
		body             any
		loggerIdentifier = fmt.Sprintf("[Request] %s", c.clientID)
	)

	if req.Body == nil {
		c.logger(ctx, loggerIdentifier, zap.Error(fmt.Errorf("failed to read request body: request body is nil")))
		return
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		c.logger(ctx, loggerIdentifier, zap.Error(fmt.Errorf("failed to read request body: %w", err)))
		return
	}
	defer req.Body.Close()

	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	contentType := req.Header.Get("Content-Type")
	isJSON := strings.Contains(strings.ToLower(contentType), "application/json")

	body = any(bodyBytes)
	if isJSON {
		body = json.RawMessage(bodyBytes)
	}

	c.logger(ctx,
		loggerIdentifier,
		zap.String("url", req.URL.String()),
		zap.Any("headers", masking.ShouldMaskStruct(req.Header)),
		zap.Any("body", masking.ShouldMaskStruct(body)),
	)
}

func (c *restClient) AfterResponse(ctx context.Context, requestTime time.Time, req *http.Request, resp *http.Response) {
	if c.logger == nil {
		return
	}

	var (
		body             any
		loggerIdentifier = fmt.Sprintf("[RESPONSE] %s", c.clientID)
	)

	if resp == nil {
		c.logger(ctx, loggerIdentifier, zap.Error(fmt.Errorf("failed to read response body: response is nil")))
		return
	}

	if resp.Body == nil {
		c.logger(ctx, loggerIdentifier, zap.Error(fmt.Errorf("failed to read response body: response body is nil")))
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger(ctx, loggerIdentifier, zap.Error(fmt.Errorf("failed to read response body: %w", err)))
		return
	}
	defer resp.Body.Close()

	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	contentType := resp.Header.Get("Content-Type")
	isJSON := strings.Contains(strings.ToLower(contentType), "application/json")

	body = any(bodyBytes)
	if isJSON {
		body = json.RawMessage(bodyBytes)
	}

	c.logger(ctx,
		loggerIdentifier,
		zap.String("url", req.URL.String()),
		zap.Int64("time_duration", time.Since(requestTime).Milliseconds()),
		zap.Int("http_status", resp.StatusCode),
		zap.Any("headers", masking.ShouldMaskStruct(resp.Header)),
		zap.Any("body", masking.ShouldMaskStruct(body)),
	)
}

func (c *restClient) ParseResponseBody(resp *http.Response, targetStruct any) error {
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if err := json.NewDecoder(resp.Body).Decode(&targetStruct); err != nil {
		return err
	}

	return nil
}

func (c *restClient) prepareRequest(ctx context.Context, method, path string, payload *RequestPayload) (*http.Request, error) {
	var (
		err           error
		reader        io.Reader
		requestHeader = c.defaultHeaders.Clone()
		fullURL       = strings.TrimSuffix(c.baseURL, "/") + "/" + strings.TrimPrefix(path, "/")
	)

	if payload != nil {
		reader, err = payload.parseBodyToReader()
		if err != nil {
			return nil, err
		}

		if len(payload.headers) > 0 {
			for key, val := range payload.headers {
				for _, v := range val {
					requestHeader.Add(key, v)
				}
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reader)
	if err != nil {
		return nil, err
	}

	req.Header = requestHeader

	return req, nil
}
