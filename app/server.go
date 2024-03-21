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

	lines := strings.Split(request, CRLF)
	startLine := strings.Split(lines[0], " ")
	method := startLine[0]
	path := startLine[1]
	versoin := startLine[2]

	fmt.Printf("method: %s, path: %s, version: %s\n", method, path, versoin)

	status := 200
	msg := "OK"
	responseHeader := "" //Content-Length: 0
	if path != "/" {
		status = 404
		msg = "Not Found"
	}
	responseStartLine := fmt.Sprintf("%s %d %s", versoin, status, msg)
	response := []byte(fmt.Sprintf("%s%s%s%s", responseStartLine, CRLF, CRLF, responseHeader))
	conn.Write(response)
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
}
