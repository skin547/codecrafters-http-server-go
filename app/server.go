package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
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
	case strings.HasPrefix(req.path, "/files/"):
		splitedPath := strings.Split(req.path, "/files/")
		fileName := splitedPath[1]
		// check if file exist
		filePath := fmt.Sprintf("files/%s", fileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			res.statusCode = 404
			res.statusMsg = "Not Found"
			break
		}

		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println(err)
			res.statusCode = 500
			res.statusMsg = "Internal Server Error"
			break
		}
		defer file.Close()
		// write to response body as string
		data := make([]byte, 1024)
		size, err := file.Read(data)
		if err != nil {
			fmt.Println(err)
			res.statusCode = 500
			res.statusMsg = "Internal Server Error"
			break
		}
		res.body = string(data[:size])
		res.headers["Content-Type"] = "text/plain"
	default:
		res.statusCode = 404
		res.statusMsg = "Not Found"
	}

	str := SerializeResponse(res)
	conn.Write([]byte(str))
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