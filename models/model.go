package models

import (
	"net/http"
)

type RequestInfo struct {
	RequestID string      `json:"requestId"`
	Path      string      `json:"path"`
	Query     interface{} `json:"query"`
	Header    http.Header `json:"header"`
	Body      []byte      `json:"body"`
}

type ResponseInfo struct {
	RequestID  string      `json:"requestId"`
	StatusCode int         `json:"statusCode"`
	Header     http.Header `json:"header"`
	Body       []byte      `json:"body"`
}
