package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"
)

func main() {
	fmt.Println("Initializating server...")

	connection := NewConnection("tcp", "localhost", 4221)
	listener, err := connection.connect()

	if err != nil {
		fmt.Printf("Error establishing a connection: %s\n", err.Error())
		os.Exit(1)
	}

	defer listener.Close()

	setupRoutes(connection)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go connection.handle(conn)
	}
}

func setupRoutes(c *Connection) {
	c.Use("GET", "/", func(request *HTTPRequest) {
		response := NewResponse(200, nil, nil)
		response.send(request.Conn)
	})
	c.Use("GET", "/index.html", func(request *HTTPRequest) {
		response := NewResponse(200, nil, nil)
		response.send(request.Conn)
	})
	c.Use("GET", "/echo", func(request *HTTPRequest) {
		resp := strings.TrimPrefix(request.URI, "/echo/")
		encodings := strings.Split(request.Headers["accept-encoding"], ", ")

		if slices.Contains(encodings, "gzip") {
			compressed, err := compressString(resp)
			if err != nil {
				fmt.Printf("Error with compression: %s\n", err.Error())
				response := NewResponse(400, nil, nil)
				response.send(request.Conn)
				return
			}
			response := NewResponse(200, []Header{
				{"Content-Encoding", "gzip"},
				ContentTypeHeader("text/plain"),
				ContentLengthHeader(len(compressed)),
			}, compressed)
			response.send(request.Conn)
			return
		}

		response := NewResponse(200, []Header{
			ContentTypeHeader("text/plain"),
			ContentLengthHeader(len(resp)),
		}, []byte(resp))
		response.send(request.Conn)
	})
	c.Use("GET", "/user-agent", func(request *HTTPRequest) {
		userAgent := request.Headers["user-agent"]
		response := NewResponse(200, []Header{
			ContentTypeHeader("text/plain"),
			ContentLengthHeader(len(userAgent)),
		}, []byte(userAgent))
		response.send(request.Conn)
	})
	c.Use("GET", "/files", func(request *HTTPRequest) {
		dir := os.Args[2]
		filename := strings.TrimPrefix(request.URI, "/files/")
		filepath := path.Join(dir, filename)

		content, err := os.ReadFile(filepath)
		if err == nil {
			response := NewResponse(200, []Header{
				ContentTypeHeader("application/octet-stream"),
				ContentLengthHeader(len(content)),
			}, content)
			response.send(request.Conn)
		} else {
			response := NewResponse(404, nil, nil)
			response.send(request.Conn)
		}

	})
	c.Use("POST", "/files", func(request *HTTPRequest) {
		dir := os.Args[2]
		filename := strings.TrimPrefix(request.URI, "/files/")
		filepath := path.Join(dir, filename)

		file, err := os.Create(filepath)
		defer file.Close()
		if err != nil {
			fmt.Printf("Error creating file: %s\n", err.Error())
			response := NewResponse(400, nil, nil)
			response.send(request.Conn)
		}

		data := []byte(bytes.Trim([]byte(request.Body), "\x00"))
		_, err = file.Write(data)
		if err != nil {
			fmt.Printf("Error creating the file: %s\n", err.Error())
			response := NewResponse(404, nil, nil)
			response.send(request.Conn)
		} else {
			response := NewResponse(201, nil, nil)
			response.send(request.Conn)
		}
	})
}
