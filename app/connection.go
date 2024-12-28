package main

import (
	"fmt"
	"net"
	"strings"
)

type Route struct {
	Method  Methods
	URI     string
	Handler func(request *HTTPRequest, response *Response)
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

func (c *Connection) Use(method Methods, uri string, handler func(request *HTTPRequest, response *Response)) {
	c.Routes = append(c.Routes, &Route{
		Method:  method,
		URI:     uri,
		Handler: handler,
	})
}

func (c *Connection) ProcessRoutes(request *HTTPRequest, conn net.Conn) {
	foundRoute := false
	for _, route := range c.Routes {
		if route.match(request) {
			foundRoute = true
			response := NewResponse(conn)
			route.Handler(request, response)
			return
		}

	}

	if !foundRoute {
		response := NewResponse(conn)
		response.Status(404).Send()
	}
}

func (c *Connection) handle(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	requestTimeout, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading data: ", err.Error())
		response := NewResponse(conn)
		response.Status(401).Body([]byte("Bad Request")).Send()
		return
	}

	request, err := ParseRequest(buf, requestTimeout, conn)

	if err != nil {
		fmt.Printf("Error with request: %s\n", err.Error())
		response := NewResponse(conn)
		response.Status(400).Body([]byte("Bad Request")).Send()
		return
	}

	c.ProcessRoutes(request, conn)
}
