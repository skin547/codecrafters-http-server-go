package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/http-server-starter-go/internal"
)


func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	var public string
	flag.StringVar(&public, "directory", "./public", "files directory")
	flag.Parse()
	// if directory not exist, create one
	_, err := os.Stat(public)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(public, 0755)
			if err != nil {
				fmt.Println("Error creating directory: ", err.Error())
				os.Exit(1)
			}
		} else {
			fmt.Println("Error checking directory: ", err.Error())
			os.Exit(1)
		}
	}

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()
	fileStorage := internal.NewFileStorage(public)
	httpServer := internal.NewHttpServer(fileStorage)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			conn.Close()
			continue
		}
		go httpServer.Handle(conn)
	}
}
