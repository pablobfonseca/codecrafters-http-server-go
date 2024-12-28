package main

type HTTPResponse string

const (
	OK         HTTPResponse = "HTTP/1.1 200 OK\r\n\r\n"
	NotFound   HTTPResponse = "HTTP/1.1 404 Not Found\r\n\r\n"
	Created    HTTPResponse = "HTTP/1.1 201 Created\r\n\r\n"
	BadRequest HTTPResponse = "HTTP/1.1 400 Bad Request\r\n\r\n"
)
