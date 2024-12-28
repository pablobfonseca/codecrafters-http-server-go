package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
)

type Route struct {
	Method  Methods
	URI     string
	Handler func(request *HTTPRequest) HTTPResponse
}

func (r *Route) match(request *HTTPRequest) bool {
	if request.Method != r.Method {
		return false
	}

	if len(r.URI) > 1 {
		_, found := strings.CutPrefix(request.URI, r.URI)
		return found
	}

	return request.URI == r.URI
}

type Connection struct {
	Port     int
	Address  string
	Protocol string
	Listener net.Listener
	Routes   []*Route
}

func NewConnection(protocol, address string, port int) *Connection {
	return &Connection{
		Port:     port,
		Address:  address,
		Protocol: protocol,
	}
}

func (c *Connection) connect() (net.Listener, error) {
	listener, err := net.Listen(c.Protocol, fmt.Sprintf("%s:%d", c.Address, c.Port))
	c.Listener = listener

	if err != nil {
		return nil, err
	}

	return c.Listener, nil
}

func (c *Connection) Use(method Methods, uri string, handler func(request *HTTPRequest) HTTPResponse) {
	c.Routes = append(c.Routes, &Route{
		Method:  method,
		URI:     uri,
		Handler: handler,
	})
}

func (c *Connection) ProcessRoutes(request *HTTPRequest, conn net.Conn) {
	foundRoute := false
	for _, route := range c.Routes {
		match := route.match(request)
		if match {
			foundRoute = true
			resp := route.Handler(request)

			conn.Write([]byte(resp))
			return
		}

	}

	if !foundRoute {
		conn.Write([]byte(NotFound))
	}
}

func (c *Connection) handle(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	requestTimeout, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading data: ", err.Error())
	}

	request, err := ParseRequest(buf, requestTimeout)

	if err != nil {
		fmt.Printf("Error with request: %s\n", err.Error())
	}

	c.Use("GET", "/", func(request *HTTPRequest) HTTPResponse {
		return OK
	})
	c.Use("GET", "/index.html", func(request *HTTPRequest) HTTPResponse {
		return OK
	})
	c.Use("GET", "/echo", func(request *HTTPRequest) HTTPResponse {
		resp := strings.TrimPrefix(request.URI, "/echo/")
		encoding := request.Headers["accept-encoding"]

		if encoding == "gzip" {
			compressed, err := compressString(resp)
			if err != nil {
				fmt.Printf("Error with compression: %s\n", err.Error())
				return BadRequest
			}
			return HTTPResponse(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(compressed), string(compressed)))
		}

		return HTTPResponse(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(resp), string(resp)))
	})
	c.Use("GET", "/user-agent", func(request *HTTPRequest) HTTPResponse {
		userAgent := request.Headers["user-agent"]
		return HTTPResponse(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent))
	})
	c.Use("GET", "/files", func(request *HTTPRequest) HTTPResponse {
		dir := os.Args[2]
		filename := strings.TrimPrefix(request.URI, "/files/")
		filepath := path.Join(dir, filename)

		content, err := os.ReadFile(filepath)
		if err == nil {
			return HTTPResponse(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(content), content))
		} else {
			return NotFound
		}

	})
	c.Use("POST", "/files", func(request *HTTPRequest) HTTPResponse {
		dir := os.Args[2]
		filename := strings.TrimPrefix(request.URI, "/files/")
		filepath := path.Join(dir, filename)

		file, err := os.Create(filepath)
		defer file.Close()
		if err != nil {
			fmt.Printf("Error creating file: %s\n", err.Error())
			return BadRequest
		}

		data := []byte(bytes.Trim([]byte(request.Body), "\x00"))
		_, err = file.Write(data)
		if err != nil {
			fmt.Printf("Error creating the file: %s\n", err.Error())
			return NotFound
		} else {
			return Created
		}
	})

	c.ProcessRoutes(request, conn)
}

func compressString(s string) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	defer gzWriter.Close()

	_, err := gzWriter.Write([]byte(s))
	if err != nil {
		return nil, err
	}

	err = gzWriter.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
