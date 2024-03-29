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

type Status struct {
	OK                    string
	CREATED               string
	NOT_FOUND             string
	INTERNAL_SERVER_ERROR string
}

var StatusCode = Status{
	OK:                    "HTTP/1.1 200 OK",
	CREATED:               "HTTP/1.1 201 CREATED",
	NOT_FOUND:             "HTTP/1.1 404 NOT_FOUND",
	INTERNAL_SERVER_ERROR: "HTTP/1.1 500 INTERNAL_SERVER_ERROR",
}

type Content struct {
	text string
	file string
}

var ContentType = Content{
	text: "\r\nContent-Type: text/plain\r\n",
	file: "\r\nContent-Type: application/octet-stream\r\n",
}

var directory string

func main() {
	flag.StringVar(&directory, "directory", "", "directory to read and write files")
	flag.Parse()

	host, port := "127.0.0.1", "4221"
	address := host + ":" + port

	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Failed to bind to port", port)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Listening on http://" + listener.Addr().String())

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go router(conn)
	}
}

type Request struct {
	method     string
	path       string
	user_agent string
	body       string
}

func router(conn net.Conn) {
	request_bytes := make([]byte, 1024)
	n, err := conn.Read(request_bytes)
	if err != nil {
		log.Fatal(err)
	}
	request := string(request_bytes)
	lines := strings.Split(request[:n], "\r\n")

	start_line := lines[0]
	start_line_parts := strings.Split(start_line, " ")
	method := start_line_parts[0]
	path, err := url.QueryUnescape(start_line_parts[1])
	if err != nil {
		log.Fatal(err)
	}

	var user_agent string
	for _, line := range lines {
		if strings.HasPrefix(line, "User-Agent: ") {
			user_agent, _ = strings.CutPrefix(line, "User-Agent: ")
			break
		}
	}

	body := lines[len(lines)-1]

	r := Request{
		method,
		path,
		user_agent,
		body,
	}

	var response string

	if path == "/" && method == "GET" {
		response = get_health()
	}
	if path == "/user-agent" && method == "GET" {
		response = get_user_agent(r)
	}
	if strings.HasPrefix(path, "/files/") && method == "GET" {
		response = get_file(r)
	}
	if strings.HasPrefix(path, "/files/") && method == "POST" {
		response = post_file(r)
	}
	if response == "" {
		response = StatusCode.NOT_FOUND
	}

	_, err = conn.Write([]byte(response + "\r\n\r\n"))
	if err != nil {
		fmt.Println(err)
	}

	conn.Close()
}

func get_health() string {
	return StatusCode.OK
}

func get_user_agent(r Request) string {
	content_length := strconv.Itoa(len(r.user_agent))

	return StatusCode.OK +
		ContentType.text +
		"Content-Length: " + content_length +
		"\r\n\r\n" +
		r.user_agent
}

func get_file(r Request) string {
	file_name, _ := strings.CutPrefix(r.path, "/files/")
	file, err := os.Open(directory + "/" + file_name)
	if err != nil {
		return StatusCode.NOT_FOUND
	}

	file_content := make([]byte, 1024)
	n, err := file.Read(file_content)
	if err != nil {
		fmt.Println(err)
		return StatusCode.INTERNAL_SERVER_ERROR
	}

	content_length := strconv.Itoa(n)
	body := string(file_content[:n])

	file.Close()
	return StatusCode.OK +
		ContentType.file +
		"Content-Length: " + content_length +
		"\r\n\r\n" +
		body
}

func post_file(r Request) string {
	file_name, _ := strings.CutPrefix(r.path, "/files/")

	err := os.WriteFile(directory+"/"+file_name, []byte(r.body), 0666)
	if err != nil {
		fmt.Println(err)
		return StatusCode.INTERNAL_SERVER_ERROR
	}
	return StatusCode.CREATED
}
