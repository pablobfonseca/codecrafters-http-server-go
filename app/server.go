package main

import (
	"bufio"
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

	scanner := bufio.NewScanner(strings.NewReader(string(buf[:])))

	var requestLine string

	if scanner.Scan() {
		requestLine = scanner.Text()
	}

	headers := parseHeaders(scanner)

	url := strings.Split(requestLine, " ")[1]

	if url == "/index.html" || url == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if strings.HasPrefix(url, "/echo/") {
		resp, _ := strings.CutPrefix(url, "/echo/")

		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(resp), string(resp))))
	} else if url == "/user-agent" {
		userAgent := headers["user-agent"]
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
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
