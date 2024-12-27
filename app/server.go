package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading data: ", err.Error())
	}

	requestData := bytes.Split(buf, []byte("\r\n"))
	requestLine := string(requestData[0])
	url := strings.Split(requestLine, " ")[1]

	if url == "/index.html" || url == "/" {
		success(conn)
	} else if strings.HasPrefix(url, "/echo/") {
		resp, _ := strings.CutPrefix(url, "/echo/")

		successWithResponse(conn, []byte(resp))
	} else {
		notFound(conn)
	}
}

func success(conn net.Conn) {
	_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	if err != nil {
		fmt.Println("Error sending data: ", err.Error())
	}
}

func successWithResponse(conn net.Conn, responseBody []byte) {
	_, err := conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(responseBody), string(responseBody))))

	if err != nil {
		fmt.Println("Error sending data: ", err.Error())
	}
}

func notFound(conn net.Conn) {
	_, err := conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	if err != nil {
		fmt.Println("Error sending data: ", err.Error())
	}
}
