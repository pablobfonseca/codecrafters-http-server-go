package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
)

type Methods string

const (
	GET    Methods = "GET"
	POST   Methods = "POST"
	PUT    Methods = "PUT"
	DELETE Methods = "DELETE"
	PATCH  Methods = "PATCH"
)

type HTTPRequest struct {
	Method   Methods
	Protocol string
	URI      string
	Headers  map[string]string
	Body     []byte
	Conn     net.Conn
}

func ParseRequest(requestBuffer []byte, requestTimeout int, conn net.Conn) (*HTTPRequest, error) {
	if requestTimeout == 0 || len(requestBuffer) == 0 {
		return nil, errors.New("invalid request: timeout or empty buffer")
	}

	lines := strings.Split(string(requestBuffer), "\r\n")

	if len(lines) < 1 {
		return nil, errors.New("invalid request: missing request line")
	}

	emptyLineIndex := 0
	for i, line := range lines {
		if line == "" {
			emptyLineIndex = i
			break
		}
	}

	headersLines := lines[1:emptyLineIndex]
	bodyLines := lines[emptyLineIndex+1:]

	requestLine := lines[0]

	parts := strings.SplitN(requestLine, " ", 3)
	if len(parts) < 3 {
		return nil, errors.New("Malformed request line")
	}

	r := &HTTPRequest{
		Method:   Methods(strings.ToUpper(parts[0])),
		URI:      parts[1],
		Protocol: parts[2],
		Conn:     conn,
	}

	r.parseHeaders(headersLines)
	r.parseBody(bodyLines)

	return r, nil
}

func (r *HTTPRequest) parseBody(bodyLines []string) {
	body := strings.Join(bodyLines, "\n")
	r.Body = []byte(strings.TrimSpace(strings.ReplaceAll(body, "\x00", "")))
}

func (r *HTTPRequest) ParseJSON(target interface{}) error {
	if len(r.Body) == 0 {
		return errors.New("request body is empty")
	}

	if err := json.Unmarshal([]byte(r.Body), target); err != nil {
		return fmt.Errorf("failed to parse JSON: %w\n", err)
	}

	return nil
}

func (r *HTTPRequest) parseHeaders(headerLine []string) {
	headers := make(map[string]string)

	for _, line := range headerLine {
		if line == "" || !strings.Contains(line, ": ") {
			continue
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			headers[strings.ToLower(parts[0])] = parts[1]
		}
	}

	r.Headers = headers
}
