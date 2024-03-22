package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const CRLF = "\r\n"

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

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
			conn.Close()
		}
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	requestSize, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading connection: ", err.Error())
		conn.Close()
		return
	}
	request := string(buf[:requestSize])

	req := ParseRequest(request)
	fmt.Printf("method: %s, path: %s, query: %s, version: %s\n", req.method, req.path, req.query, req.version)
	fmt.Printf("headers: %v\n", req.headers)

	status := 200
	msg := "OK"
	responseHeader := ""
	responesBody := ""

	switch {
	case req.path == "/":

	case strings.HasPrefix(req.path, "/echo/"):
		splitedPath := strings.Split(req.path, "/echo/")
		echo := splitedPath[1]

		responseHeader += "Content-Type: text/plain"
		responseHeader += CRLF
		responseHeader += fmt.Sprintf("Content-Length: %d", len(echo))
		responesBody = echo
	case strings.HasPrefix(req.path, "/user-agent"):
		userAgent, exist := req.headers["User-Agent"]
		if !exist {
			userAgent = "Unknown"
		}
		responesBody = userAgent
		responseHeader += "Content-Type: text/plain"
		responseHeader += CRLF
		responseHeader += fmt.Sprintf("Content-Length: %d", len(responesBody))
	default:
		status = 404
		msg = "Not Found"
	}
	statusLine := fmt.Sprintf("%s %d %s", req.version, status, msg)
	response := []byte(fmt.Sprintf("%s%s%s%s%s%s%s", statusLine, CRLF, responseHeader, CRLF, CRLF, responesBody, CRLF))
	conn.Write(response)
}

// path contain path and query string
func splitPath(path string) (string, string) {
	splited := strings.Split(path, "?")
	if len(splited) == 1 {
		return splited[0], ""
	}
	return splited[0], splited[1]
}

type Request struct {
	method  string
	path    string
	query   string
	version string
	headers map[string]string
}

func ParseRequest(request string) Request {
	lines := strings.Split(request, CRLF)
	requestLine := strings.Split(lines[0], " ")
	method := requestLine[0]
	headers := make(map[string]string)

	if len(lines) > 1 {
		for _, line := range lines[1:] {
			splited := strings.Split(line, ": ")
			if len(splited) == 2 {
				key := splited[0]
				value := splited[1]
				headers[key] = value
			}
		}
	}

	paths, query := splitPath(requestLine[1])
	versoin := requestLine[2]
	return Request{method, paths, query, versoin, headers}
}
