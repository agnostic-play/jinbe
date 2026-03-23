package http_client

import (
	"net/http"
)

type PayloadSetterFn func(values *RequestPayload)
type HeaderSetterFn func(values *http.Header)

func WithFormField(key, value string) PayloadSetterFn {
	return func(payload *RequestPayload) {
		payload.formBody.Set(key, value)
	}
}

func WithHeader(header http.Header) PayloadSetterFn {
	return func(payload *RequestPayload) {
		payload.headers = header
	}
}

func WithHeaders(fn ...HeaderSetterFn) PayloadSetterFn {
	return func(payload *RequestPayload) {
		for _, setter := range fn {
			setter(&payload.headers)
		}
	}
}

func SetHeaderValue(key, value string) HeaderSetterFn {
	return func(h *http.Header) {
		h.Set(key, value)
	}
}

func WithJsonBody(jsonBody any) PayloadSetterFn {
	return func(payload *RequestPayload) {
		payload.body = jsonBody
	}
}
