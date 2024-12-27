package main

import (
	"bufio"
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
	Method   string
	Protocol string
	URI      string
	Headers  map[string]string
}

func ParseRequest(requestBuffer []byte, requestTimeout int) (*HTTPRequest, error) {
	if requestTimeout == 0 || len(requestBuffer) == 0 {
		return nil, errors.New("Invalid request")
	}
	scanner := bufio.NewScanner(strings.NewReader(string(requestBuffer[:])))
	var requestLine string
	if scanner.Scan() {
		requestLine = scanner.Text()
	}

	r := &HTTPRequest{}
	r.Method = strings.Split(requestLine, " ")[0]
	r.URI = strings.Split(requestLine, " ")[1]
	r.Protocol = strings.Split(requestLine, " ")[2]
	r.Headers = parseHeaders(scanner)

	return r, nil
}

func parseHeaders(scanner *bufio.Scanner) map[string]string {
	headers := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			break
		}

		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			key, value := parts[0], parts[1]
			headers[strings.ToLower(key)] = value
		}
	}

	return headers
}
