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
	status  Status
	body    []byte
	conn    net.Conn
	Headers map[Headers]string
}

func (r *Response) Status(statusCode int) *Response {
	status, valid := STATUS_CODES[statusCode]

	if !valid {
		status = "Unknown Status"
	}

	r.status = Status{
		Code:   statusCode,
		Status: status,
	}

	return r
}

func NewResponse(conn net.Conn) *Response {
	return &Response{
		Headers: make(map[Headers]string),
		conn:    conn,
	}
}

func (r *Response) Body(data []byte) *Response {
	r.body = data
	return r
}

func (r *Response) AddHeader(name Headers, value string) {
	r.Headers[name] = value
}

func (r *Response) GetHeader(name Headers) (string, bool) {
	value, exists := r.Headers[name]
	return value, exists
}

func (r *Response) Send() {
	r.conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.status.Code, r.status.Status)))

	for name, value := range r.Headers {
		r.conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", name, value)))
	}
	r.conn.Write([]byte("\r\n"))
	r.conn.Write(r.body)
}

func (r *Response) AddContentTypeHeader(value string) {
	r.Headers["Content-Type"] = value
}

func (r *Response) AddContentLengthHeader(length int) {
	r.Headers["Content-Length"] = fmt.Sprintf("%d", length)
}
