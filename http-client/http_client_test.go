package http_client

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gojek/heimdall/v7/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestNewRestClient tests the creation of a new REST client
func TestNewRestClient(t *testing.T) {
	client := httpclient.NewClient()
	loggerFn := func(ctx context.Context, identifier string, objects ...zap.Field) {
		// Test logger
	}

	rc := NewRestClient("test-client", "https://api.example.com", client, loggerFn)

	assert.NotNil(t, rc)
	restClient, ok := rc.(*restClient)
	require.True(t, ok)
	assert.Equal(t, "test-client", restClient.clientID)
	assert.Equal(t, "https://api.example.com", restClient.baseURL)
	assert.NotNil(t, restClient.client)
	assert.NotNil(t, restClient.logger)
}

// TestNewRestClient_NilLogger tests client creation with nil logger
func TestNewRestClient_NilLogger(t *testing.T) {
	client := httpclient.NewClient()
	rc := NewRestClient("test-client", "https://api.example.com", client, nil)

	assert.NotNil(t, rc)
	restClient, ok := rc.(*restClient)
	require.True(t, ok)
	assert.NotNil(t, restClient.logger, "Logger should be set to default function")
}

// TestRestClient_Get tests GET requests
func TestRestClient_Get(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/test/path", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	client := httpclient.NewClient()
	rc := NewRestClient("test-client", server.URL, client, nil)

	resp, err := rc.Get(context.Background(), "test/path", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "success")
}

// TestRestClient_Post tests POST requests
func TestRestClient_Post(t *testing.T) {
	expectedBody := map[string]string{"key": "value"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/test/path", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), "key")

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"created": true}`))
	}))
	defer server.Close()

	client := httpclient.NewClient()
	rc := NewRestClient("test-client", server.URL, client, nil)

	payload := NewRequestPayload(JsonPayload, WithJsonBody(expectedBody))
	resp, err := rc.Post(context.Background(), "test/path", payload)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// TestRestClient_Put tests PUT requests
func TestRestClient_Put(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := httpclient.NewClient()
	rc := NewRestClient("test-client", server.URL, client, nil)

	resp, err := rc.Put(context.Background(), "test/path", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestRestClient_Patch tests PATCH requests
func TestRestClient_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := httpclient.NewClient()
	rc := NewRestClient("test-client", server.URL, client, nil)

	resp, err := rc.Patch(context.Background(), "test/path", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestRestClient_Delete tests DELETE requests
func TestRestClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := httpclient.NewClient()
	rc := NewRestClient("test-client", server.URL, client, nil)

	resp, err := rc.Delete(context.Background(), "test/path", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestRestClient_URLConstruction tests URL building with various paths
func TestRestClient_URLConstruction(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		path     string
		expected string
	}{
		{
			name:     "base without trailing slash, path without leading slash",
			baseURL:  "https://api.example.com",
			path:     "users",
			expected: "/users",
		},
		{
			name:     "base with trailing slash, path without leading slash",
			baseURL:  "https://api.example.com/",
			path:     "users",
			expected: "/users",
		},
		{
			name:     "base without trailing slash, path with leading slash",
			baseURL:  "https://api.example.com",
			path:     "/users",
			expected: "/users",
		},
		{
			name:     "base with trailing slash, path with leading slash",
			baseURL:  "https://api.example.com/",
			path:     "/users",
			expected: "/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expected, r.URL.Path)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := httpclient.NewClient()
			rc := NewRestClient("test-client", server.URL, client, nil)

			resp, err := rc.Get(context.Background(), tt.path, nil)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

// TestRestClient_ErrorHandling tests error scenarios
func TestRestClient_ErrorHandling(t *testing.T) {
	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal server error"}`))
		}))
		defer server.Close()

		client := httpclient.NewClient()
		rc := NewRestClient("test-client", server.URL, client, nil)

		resp, err := rc.Get(context.Background(), "test/path", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("network error", func(t *testing.T) {
		client := httpclient.NewClient(httpclient.WithHTTPTimeout(100 * time.Millisecond))
		rc := NewRestClient("test-client", "http://localhost:99999", client, nil)

		_, err := rc.Get(context.Background(), "test/path", nil)
		assert.Error(t, err)
	})
}

// TestRestClient_ParseResponseBody tests response body parsing
func TestRestClient_ParseResponseBody(t *testing.T) {
	type TestResponse struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}

	t.Run("successful parsing", func(t *testing.T) {
		body := bytes.NewReader([]byte(`{"message": "test", "code": 200}`))
		resp := &http.Response{
			Body: io.NopCloser(body),
		}

		client := httpclient.NewClient()
		rc := NewRestClient("test-client", "http://example.com", client, nil)

		var result TestResponse
		err := rc.ParseResponseBody(resp, &result)

		require.NoError(t, err)
		assert.Equal(t, "test", result.Message)
		assert.Equal(t, 200, result.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		body := bytes.NewReader([]byte(`{invalid json}`))
		resp := &http.Response{
			Body: io.NopCloser(body),
		}

		client := httpclient.NewClient()
		rc := NewRestClient("test-client", "http://example.com", client, nil)

		var result TestResponse
		err := rc.ParseResponseBody(resp, &result)

		assert.Error(t, err)
	})
}

// TestRestClient_WithHeaders tests request with custom headers
func TestRestClient_WithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := httpclient.NewClient()
	rc := NewRestClient("test-client", server.URL, client, nil)

	payload := NewRequestPayload(
		JsonPayload,
		WithHeaders(
			SetHeaderValue("Authorization", "Bearer token123"),
		),
	)

	resp, err := rc.Get(context.Background(), "test/path", payload)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestRestClient_FormPayload tests form-encoded requests
func TestRestClient_FormPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		err := r.ParseForm()
		require.NoError(t, err)
		assert.Equal(t, "value1", r.Form.Get("field1"))
		assert.Equal(t, "value2", r.Form.Get("field2"))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := httpclient.NewClient()
	rc := NewRestClient("test-client", server.URL, client, nil)

	payload := NewRequestPayload(
		FormPayload,
		WithFormField("field1", "value1"),
		WithFormField("field2", "value2"),
	)

	resp, err := rc.Post(context.Background(), "test/path", payload)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// BenchmarkRestClient_Get benchmarks GET requests
func BenchmarkRestClient_Get(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	client := httpclient.NewClient()
	rc := NewRestClient("bench-client", server.URL, client, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rc.Get(ctx, "test/path", nil)
	}
}
