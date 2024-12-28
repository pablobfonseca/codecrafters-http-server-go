package main

import (
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
	c.Use("GET", "/", func(request *HTTPRequest, response *Response) {
		response.Status(200).Send()
	})
	c.Use("GET", "/index.html", func(request *HTTPRequest, response *Response) {
		response.Status(200).Send()
	})
	c.Use("GET", "/echo", func(request *HTTPRequest, response *Response) {
		resp := strings.TrimPrefix(request.URI, "/echo/")
		encodings := strings.Split(request.Headers["accept-encoding"], ", ")

		if slices.Contains(encodings, "gzip") {
			response.CompressAndSend(resp)
			return
		}

		response.AddContentTypeHeader("text/plain")
		response.AddContentLengthHeader(len(resp))
		response.Status(200).Body([]byte(resp)).Send()
	})
	c.Use("GET", "/user-agent", func(request *HTTPRequest, response *Response) {
		userAgent := request.Headers["user-agent"]

		response.AddContentTypeHeader("text/plain")
		response.AddContentLengthHeader(len(userAgent))
		response.Status(200).Body([]byte(userAgent)).Send()
	})
	c.Use("GET", "/files", func(request *HTTPRequest, response *Response) {
		dir := os.Args[2]
		filename := strings.TrimPrefix(request.URI, "/files/")
		filepath := path.Join(dir, filename)

		content, err := os.ReadFile(filepath)
		if err == nil {
			response.AddContentTypeHeader("application/octet-stream")
			response.AddContentLengthHeader(len(content))
			response.Status(200).Body(content).Send()
		} else {
			response.Status(404).Send()
		}

	})
	c.Use("POST", "/files", func(request *HTTPRequest, response *Response) {
		dir := os.Args[2]
		filename := strings.TrimPrefix(request.URI, "/files/")
		filepath := path.Join(dir, filename)

		file, err := os.Create(filepath)
		defer file.Close()
		if err != nil {
			fmt.Printf("Error creating file: %s\n", err.Error())
			response.Status(400).Send()
		}

		_, err = file.Write(request.Body)
		if err != nil {
			fmt.Printf("Error creating the file: %s\n", err.Error())
			response.Status(404).Send()
		} else {
			response.Status(201).Send()
		}
	})
	c.Use("POST", "/user", func(request *HTTPRequest, response *Response) {
		var data map[string]interface{}
		if err := request.ParseJSON(&data); err != nil {
			response.Status(400).Body([]byte("Invalid JSON")).Send()
			return
		}

		response.Status(200).Json(data)
	})
	c.Use("GET", "/status", func(request *HTTPRequest, response *Response) {
		data := map[string]string{
			"message": "ok",
		}

		response.Status(200).Json(data)
	})
}
