package http_client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type RequestPayload struct {
	requestBodyType PayloadBodyType
	headers         http.Header
	body            any
	formBody        url.Values
	jsonBody        any
	reader          io.Reader
}

func NewRequestPayload(requestType PayloadBodyType, setterFn ...PayloadSetterFn) *RequestPayload {
	reqPayload := &RequestPayload{
		requestBodyType: requestType,
		headers:         make(http.Header),
		formBody:        make(url.Values),
	}

	reqPayload.headers.Set("Content-Type", requestType.getContentType())

	for _, fn := range setterFn {
		fn(reqPayload)
	}

	return reqPayload
}

func (r RequestPayload) parseBodyToReader() (io.Reader, error) {
	var reader io.Reader

	if r.requestBodyType == FormPayload {
		r.reader = strings.NewReader(r.formBody.Encode())
		return r.reader, nil
	}

	if r.body != nil {
		jsonBody, err := json.Marshal(r.body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(jsonBody)
	}

	return reader, nil
}
