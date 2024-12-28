package main

import (
	"fmt"
	"net"
)

type HTTPResponse string
type Statuses string
type Headers string

const (
	OK         Statuses = "OK"
	NotFound   Statuses = "Not Found"
	Created    Statuses = "Created"
	BadRequest Statuses = "Bad Request"
)

var STATUS_CODES = map[int]Statuses{
	200: OK,
	404: NotFound,
	201: Created,
	400: BadRequest,
}

const (
	ContentType   Headers = "Content-Type"
	ContentLength Headers = "Content-Length"
	Accept        Headers = "Accept"
)

type Status struct {
	Code   int
	Status Statuses
}

type Header struct {
	Name  Headers
	Value string
}

type Response struct {
	Status
	Headers map[Headers]string
	Body    []byte
}

func NewResponse(statusCode int, headers []Header, body []byte) Response {
	status, valid := STATUS_CODES[statusCode]

	if !valid {
		status = "Unknown Status"
	}

	headersMap := make(map[Headers]string)
	for _, header := range headers {
		headersMap[header.Name] = header.Value
	}

	return Response{
		Status: Status{
			Code:   statusCode,
			Status: status,
		},
		Headers: headersMap,
		Body:    body,
	}
}

func (r *Response) GetHeader(name Headers) (string, bool) {
	value, exists := r.Headers[name]
	return value, exists
}

func (r *Response) send(conn net.Conn) {
	conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.Status.Code, r.Status.Status)))

	for name, value := range r.Headers {
		conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", name, value)))
	}
	conn.Write([]byte("\r\n"))
	conn.Write(r.Body)
}

func ContentTypeHeader(value string) Header {
	return Header{"Content-Type", value}
}

func ContentLengthHeader(length int) Header {
	return Header{"Content-Length", fmt.Sprintf("%d", length)}
}
