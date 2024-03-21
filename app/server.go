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

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
	}
	buf := make([]byte, 1024)
	requestSize, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
	}
	request := string(buf[:requestSize])

	req := ParseRequest(request)
	fmt.Printf("method: %s, path: %s, query: %s, version: %s\n", req.method, req.path, req.query, req.version)

	status := 200
	msg := "OK"
	responseHeader := "" //Content-Length: 0
	if paths != "/" {
		status = 404
		msg = "Not Found"
	}
	statusLine := fmt.Sprintf("%s %d %s", versoin, status, msg)
	response := []byte(fmt.Sprintf("%s%s%s%s", statusLine, CRLF, CRLF, responseHeader))
	conn.Write(response)
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
}

// path contain path and query string
func splitPath(path string) (string, string) {
	splited := strings.Split(path, "?")
	return splited[0], splited[1]
}

type Request struct {
	method  string
	path    string
	query   string
	version string
}

func ParseRequest(request string) Request {
	lines := strings.Split(request, CRLF)
	requestLine := strings.Split(lines[0], " ")
	method := requestLine[0]

	paths, query := splitPath(requestLine[1])
	versoin := requestLine[2]
	return Request{method, paths, query, versoin}
}
