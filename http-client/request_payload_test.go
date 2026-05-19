package http_client

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewRequestPayload tests payload creation
func TestNewRequestPayload(t *testing.T) {
	t.Run("JSON payload", func(t *testing.T) {
		payload := NewRequestPayload(JsonPayload)

		assert.NotNil(t, payload)
		assert.Equal(t, JsonPayload, payload.requestBodyType)
		assert.Equal(t, "application/json", payload.headers.Get("Content-Type"))
	})

	t.Run("Form payload", func(t *testing.T) {
		payload := NewRequestPayload(FormPayload)

		assert.NotNil(t, payload)
		assert.Equal(t, FormPayload, payload.requestBodyType)
		assert.Equal(t, "application/x-www-form-urlencoded", payload.headers.Get("Content-Type"))
	})
}

// TestNewRequestPayload_WithSetters tests payload creation with setters
func TestNewRequestPayload_WithSetters(t *testing.T) {
	t.Run("with JSON body", func(t *testing.T) {
		testData := map[string]string{"key": "value"}
		payload := NewRequestPayload(JsonPayload, WithJsonBody(testData))

		assert.NotNil(t, payload)
		assert.Equal(t, testData, payload.body)
	})

	t.Run("with form fields", func(t *testing.T) {
		payload := NewRequestPayload(
			FormPayload,
			WithFormField("field1", "value1"),
			WithFormField("field2", "value2"),
		)

		assert.Equal(t, "value1", payload.formBody.Get("field1"))
		assert.Equal(t, "value2", payload.formBody.Get("field2"))
	})

	t.Run("with custom headers", func(t *testing.T) {
		payload := NewRequestPayload(
			JsonPayload,
			WithHeaders(
				SetHeaderValue("Authorization", "Bearer token"),
				SetHeaderValue("X-Custom-Header", "custom-value"),
			),
		)

		assert.Equal(t, "Bearer token", payload.headers.Get("Authorization"))
		assert.Equal(t, "custom-value", payload.headers.Get("X-Custom-Header"))
	})
}

// TestRequestPayload_parseBodyToReader tests body parsing
func TestRequestPayload_parseBodyToReader(t *testing.T) {
	t.Run("JSON body", func(t *testing.T) {
		testData := map[string]string{"key": "value"}
		payload := NewRequestPayload(JsonPayload, WithJsonBody(testData))

		reader, err := payload.parseBodyToReader()

		require.NoError(t, err)
		assert.NotNil(t, reader)
	})

	t.Run("Form body", func(t *testing.T) {
		payload := NewRequestPayload(
			FormPayload,
			WithFormField("field1", "value1"),
			WithFormField("field2", "value2"),
		)

		reader, err := payload.parseBodyToReader()

		require.NoError(t, err)
		assert.NotNil(t, reader)
	})

	t.Run("Empty JSON body", func(t *testing.T) {
		payload := NewRequestPayload(JsonPayload)

		reader, err := payload.parseBodyToReader()

		require.NoError(t, err)
		assert.Nil(t, reader)
	})

	t.Run("Invalid JSON body", func(t *testing.T) {
		payload := NewRequestPayload(JsonPayload)
		// Create a circular reference that can't be marshaled
		type Circular struct {
			Self interface{}
		}
		circular := &Circular{}
		circular.Self = circular
		payload.body = circular

		_, err := payload.parseBodyToReader()

		assert.Error(t, err)
	})
}

// TestPayloadBodyType_getContentType tests content type retrieval
func TestPayloadBodyType_getContentType(t *testing.T) {
	tests := []struct {
		name        string
		payloadType PayloadBodyType
		expected    string
	}{
		{
			name:        "JSON payload",
			payloadType: JsonPayload,
			expected:    "application/json",
		},
		{
			name:        "Form payload",
			payloadType: FormPayload,
			expected:    "application/x-www-form-urlencoded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.payloadType.getContentType()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// TestWithFormField tests form field setter
func TestWithFormField(t *testing.T) {
	payload := &RequestPayload{
		formBody: make(url.Values),
	}

	setter := WithFormField("testKey", "testValue")
	setter(payload)

	assert.Equal(t, "testValue", payload.formBody.Get("testKey"))
}

// TestWithHeader tests header setter
func TestWithHeader(t *testing.T) {
	customHeaders := http.Header{
		"Authorization": []string{"Bearer token"},
		"X-Custom":      []string{"value"},
	}

	payload := &RequestPayload{
		headers: make(http.Header),
	}

	setter := WithHeader(customHeaders)
	setter(payload)

	assert.Equal(t, customHeaders, payload.headers)
}

// TestSetHeaderValue tests individual header value setter
func TestSetHeaderValue(t *testing.T) {
	headers := make(http.Header)

	setter := SetHeaderValue("Test-Header", "test-value")
	setter(&headers)

	assert.Equal(t, "test-value", headers.Get("Test-Header"))
}

// TestWithHeaders tests multiple header setters
func TestWithHeaders(t *testing.T) {
	payload := &RequestPayload{
		headers: make(map[string][]string),
	}

	setter := WithHeaders(
		SetHeaderValue("Header1", "value1"),
		SetHeaderValue("Header2", "value2"),
	)
	setter(payload)

	assert.Equal(t, "value1", payload.headers.Get("Header1"))
	assert.Equal(t, "value2", payload.headers.Get("Header2"))
}

// TestWithJsonBody tests JSON body setter
func TestWithJsonBody(t *testing.T) {
	testData := map[string]interface{}{
		"string": "value",
		"number": 123,
		"bool":   true,
	}

	payload := &RequestPayload{}

	setter := WithJsonBody(testData)
	setter(payload)

	assert.Equal(t, testData, payload.body)
}
