package main

import (
	"errors"
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
	Body     string
}

func ParseRequest(requestBuffer []byte, requestTimeout int) (*HTTPRequest, error) {
	if requestTimeout == 0 || len(requestBuffer) == 0 {
		return nil, errors.New("Invalid request")
	}

	lines := strings.Split(string(requestBuffer), "\r\n")

	requestLine := lines[0]

	headersLines := lines[1 : len(lines)-1]
	bodyLine := lines[len(lines)-1]

	parts := strings.SplitN(requestLine, " ", 3)
	if len(parts) < 3 {
		return nil, errors.New("Malformed request line")
	}

	r := &HTTPRequest{
		Method:   Methods(parts[0]),
		URI:      parts[1],
		Protocol: parts[2],
		Headers:  parseHeaders(headersLines),
		Body:     bodyLine,
	}

	return r, nil
}

func parseHeaders(headerLine []string) map[string]string {
	headers := make(map[string]string)

	for _, line := range headerLine {
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			headers[strings.ToLower(parts[0])] = parts[1]
		}

	}

	return headers
}
