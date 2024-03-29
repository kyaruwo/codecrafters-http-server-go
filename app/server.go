package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

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

	request_bytes := make([]byte, 1024)
	_, err = conn.Read(request_bytes)
	if err != nil {
		log.Fatal(err)
	}
	request := string(request_bytes)
	lines := strings.Split(request, "\r\n")
	path := strings.Split(lines[0], " ")[1]

	var response string
	if path == "/" {
		response = "HTTP/1.1 200 OK\r\n\r\n"
	} else if strings.HasPrefix(path, "/echo/") {
		body := strings.Split(path, "/")[1]
		len := strconv.Itoa(len(body))

		response = "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n" +
			"Content-Length: " + len + "\r\n\r\n" +
			strings.Split(path, "/")[2] + "\r\n\r\n"
	} else {
		response = "HTTP/1.1 404 NOT FOUND\r\n\r\n"
	}

	_, err = conn.Write([]byte(response))
	if err != nil {
		log.Fatal(err)
	}
}
