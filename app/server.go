package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var directory string

func main() {
	flag.StringVar(&directory, "directory", "", "")
	flag.Parse()

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handler(conn)
	}
}

func handler(conn net.Conn) {
	request_bytes := make([]byte, 1024)
	_, err := conn.Read(request_bytes)
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
		body, _ := strings.CutPrefix(path, "/echo/")
		len := strconv.Itoa(len(body))

		response = "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n" +
			"Content-Length: " + len + "\r\n\r\n" +
			body + "\r\n\r\n"
	} else if path == "/user-agent" {
		var user_agent string
		for _, line := range lines {
			if strings.HasPrefix(line, "User-Agent: ") {
				user_agent, _ = strings.CutPrefix(line, "User-Agent: ")
			}
		}
		len := strconv.Itoa(len(user_agent))

		response = "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n" +
			"Content-Length: " + len + "\r\n\r\n" +
			user_agent + "\r\n\r\n"
	} else if strings.HasPrefix(path, "/files/") {
		file_name, _ := strings.CutPrefix(path, "/files/")
		file_name, err = url.QueryUnescape(file_name)
		if err != nil {
			log.Fatal(err)
		}

		file, err := os.Open(directory + "/" + file_name)
		if err != nil {
			response = "HTTP/1.1 404 NOT FOUND\r\n\r\n"
		} else {
			content := make([]byte, 1024)
			bytes_length, err := file.Read(content)
			if err != nil {
				log.Fatal(err)
			}
			len := strconv.Itoa(bytes_length)
			body := string(content)

			response = "HTTP/1.1 200 OK\r\n" +
				"Content-Type: application/octet-stream\r\n" +
				"Content-Length: " + len + "\r\n\r\n" +
				body + "\r\n\r\n"
		}
		file.Close()
	} else {
		response = "HTTP/1.1 404 NOT FOUND\r\n\r\n"
	}

	_, err = conn.Write([]byte(response))
	if err != nil {
		log.Fatal(err)
	}

	conn.Close()
}
