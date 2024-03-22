package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const CRLF = "\r\n"
var public string

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

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
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			conn.Close()
			continue
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
		return
	}
	request := string(buf[:requestSize])

	req := ParseRequest(request)
	fmt.Printf("method: %s, path: %s, query: %s, version: %s\n", req.method, req.path, req.query, req.version)
	fmt.Printf("headers: %v\n", req.headers)

	res := Response{
		body: "",
		version: req.version,
		statusCode: 200,
		statusMsg: "OK",
		headers: make(map[string]string),
	}

	switch {
	case req.path == "/":
	case strings.HasPrefix(req.path, "/echo/"):
		splitedPath := strings.Split(req.path, "/echo/")
		echo := splitedPath[1]

		res.body = echo
		res.headers["Content-Type"] = "text/plain"
	case strings.HasPrefix(req.path, "/user-agent"):
		userAgent, exist := req.headers["User-Agent"]
		if !exist {
			userAgent = "Unknown"
		}
		res.body = userAgent

		res.headers["Content-Type"] = "text/plain"
	case req.method == "GET" && strings.HasPrefix(req.path, "/files/"):
		splitedPath := strings.Split(req.path, "/files/")
		fileName := splitedPath[1]
		fmt.Printf("fileName: %s\n", fileName)
		// check if file exist
		filePath := fmt.Sprintf("%s/%s", public, fileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			res.statusCode = 404
			res.statusMsg = "Not Found"
			break
		}

		data, err := readFile(filePath)
		if err != nil {
			res.statusCode = 500
			res.statusMsg = "Internal Server Error"
			break
		}
		res.body = string(data)
		res.headers["Content-Type"] = "application/octet-stream"
	case req.method == "POST" && strings.HasPrefix(req.path, "/files/"):
		splitedPath := strings.Split(req.path, "/files/")
		fileName := splitedPath[1]
		filePath := fmt.Sprintf("%s/%s", public, fileName)
		err = os.WriteFile(filePath, []byte(req.body), 0644)
		if err != nil {
			res.statusCode = 500
			res.statusMsg = "Internal Server Error"
			break
		}
		res.body = "ok"
		res.statusCode = 201
	default:
		res.statusCode = 404
		res.statusMsg = "Not Found"
	}

	str := SerializeResponse(res)
	conn.Write([]byte(str))
}

func readFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %s", err)
	}
	defer file.Close()

	data := make([]byte, 1024)
	size, err := file.Read(data)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %s", err)
	}
	return data[:size], nil
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
	body    string
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
	version := requestLine[2]
	body := ""
	if method == "POST" {
		body := lines[len(lines)-1]
		headers["Content-Length"] = strconv.Itoa(len(body))
	}
	return Request{method, paths, query, version, headers, body}
}

type Response struct {
	version string
	statusCode int
	statusMsg string
	headers map[string]string
	body   string
}

func SerializeResponse(res Response) string {
	res.headers["Content-Length"] = strconv.Itoa(len(res.body))
	headers := ""

	for key, value := range res.headers {
		headers += fmt.Sprintf("%s: %s%s", key, value, CRLF)
	}
	return fmt.Sprintf("%s %d %s%s%s%s%s", res.version, res.statusCode, res.statusMsg, CRLF, headers, CRLF, res.body)
}