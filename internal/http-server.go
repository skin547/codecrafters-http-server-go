package internal

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const CRLF = "\r\n"

type HttpServer struct {
	storage Storage
}

func NewHttpServer(storage Storage) *HttpServer {
	return &HttpServer{
		storage: storage,
	}
}

func (h *HttpServer) Handle(conn net.Conn) {
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
	fmt.Printf("body: %v\n", req.body)

	res := Response{
		body:       "",
		version:    "HTTP/1.1",
		statusCode: 200,
		statusMsg:  "OK",
		headers:    make(map[string]string),
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

		data, err := h.storage.read(fileName)
		if err != nil {
			switch err.(type) {
			case *NotFoundError:
				res.statusCode = 404
				res.statusMsg = "Not Found"
			case *InternalServerError:
			default:
				res.statusCode = 500
				res.statusMsg = "Internal Server Error"
			}
			break
		}
		res.body = string(data)
		res.headers["Content-Type"] = "application/octet-stream"
	case req.method == "POST" && strings.HasPrefix(req.path, "/files/"):
		splitedPath := strings.Split(req.path, "/files/")
		fileName := splitedPath[1]

		err = h.storage.write(fileName, []byte(req.body))
		if err != nil {
			switch err.(type) {
			case *InternalServerError:
			default:
				res.statusCode = 500
				res.statusMsg = "Internal Server Error"
			}
			break
		}
		res.body = "ok"
		res.statusCode = 201
	default:
		res.statusCode = 404
		res.statusMsg = "Not Found"
	}

	conn.Write([]byte(res.Serialize()))
}

type Request struct {
	method  string
	path    string
	query   string
	version string
	headers map[string]string
	body    string
}

type Response struct {
	version    string
	statusCode int
	statusMsg  string
	headers    map[string]string
	body       string
}

func splitPathAndQuery(path string) (string, string) {
	splited := strings.Split(path, "?")
	if len(splited) == 1 {
		return splited[0], ""
	}
	return splited[0], splited[1]
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

	paths, query := splitPathAndQuery(requestLine[1])
	version := requestLine[2]
	body := ""
	if method == "POST" {
		body = lines[len(lines)-1]
		headers["Content-Length"] = strconv.Itoa(len(body))
	}
	return Request{method, paths, query, version, headers, body}
}

func (r *Response) Serialize() string {
	r.headers["Content-Length"] = strconv.Itoa(len(r.body))
	headers := ""

	for key, value := range r.headers {
		headers += fmt.Sprintf("%s: %s%s", key, value, CRLF)
	}
	return fmt.Sprintf("%s %d %s%s%s%s%s", r.version, r.statusCode, r.statusMsg, CRLF, headers, CRLF, r.body)
}
