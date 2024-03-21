package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()

	responseStatus := "HTTP/1.1 200 OK"
	responseHeader := "" //Content-Length: 0
	response := []byte(fmt.Sprintf("%s\r\n%s\r\n", responseStatus, responseHeader))
	conn.Write(response)
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
}
