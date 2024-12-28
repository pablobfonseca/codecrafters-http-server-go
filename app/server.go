package main

import (
	"fmt"
	"net"
	"os"
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

	if request.URI == "/index.html" || request.URI == "/" {
		conn.Write([]byte(OK))
	} else if strings.HasPrefix(request.URI, "/echo/") {
		resp, _ := strings.CutPrefix(request.URI, "/echo/")

		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(resp), string(resp))))
	} else if request.URI == "/user-agent" {
		userAgent := request.Headers["user-agent"]
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)))
	} else if strings.HasPrefix(request.URI, "/files/") {
		dir := os.Args[2]
		filename, _ := strings.CutPrefix(request.URI, "/files/")

		content, err := os.ReadFile(fmt.Sprintf("%s%s", dir, filename))
		if err == nil {
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(content), content)))
		} else {
			conn.Write([]byte(NotFound))
		}

	} else {
		conn.Write([]byte(NotFound))
	}
}
