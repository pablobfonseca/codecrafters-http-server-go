package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
)

func main() {
	fmt.Println("Initializating server...")

	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
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

	var response HTTPResponse
	if request.URI == "/index.html" || request.URI == "/" {
		response = OK
	} else if strings.HasPrefix(request.URI, "/echo/") {
		resp := strings.TrimPrefix(request.URI, "/echo/")

		response = HTTPResponse(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(resp), string(resp)))
	} else if request.URI == "/user-agent" {
		userAgent := request.Headers["user-agent"]
		response = HTTPResponse(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent))
	} else if strings.HasPrefix(request.URI, "/files/") {
		dir := os.Args[2]
		filename := strings.TrimPrefix(request.URI, "/files/")
		filepath := path.Join(dir, filename)

		if request.Method == "GET" {
			content, err := os.ReadFile(filepath)
			if err == nil {
				header := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n", len(content))
				conn.Write([]byte(header))
				conn.Write(content)
				return
			} else {
				response = NotFound
			}
		} else if request.Method == "POST" {
			file, err := os.Create(filepath)
			defer file.Close()
			if err != nil {
				fmt.Printf("Error creating file: %s\n", err.Error())
				return
			}

			data := []byte(bytes.Trim([]byte(request.Body), "\x00"))
			_, err = file.Write(data)
			if err != nil {
				fmt.Printf("Error creating the file: %s\n", err.Error())
				response = NotFound
			} else {
				response = Created
			}
		}
	} else {
		response = NotFound
	}

	conn.Write([]byte(response))
}
