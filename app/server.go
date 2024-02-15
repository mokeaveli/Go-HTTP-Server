package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

func main() {

	if len(os.Args) == 3 && os.Args[1] == "--directory" {
		dir := os.Args[2]
		if err := os.Chdir(dir); err != nil {
			log.Fatal(err.Error())
		}
	}

	l, err := net.Listen("tcp", "0.0.0.0:4221")

	if err != nil {
		log.Fatal(err.Error())
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err.Error())
		}

		go func(conn net.Conn) {
			defer conn.Close()
			buf := make([]byte, 1024)
			for {
				size, err := conn.Read(buf)
				if err != nil {
					log.Fatal(err.Error())
				}

				r, err := ParseRequest(buf[:size])
				if err != nil {
					log.Fatal(err.Error())
				}

				fmt.Println(r)
				if err := HandleFunc(CreateResponseWriter(conn, r), r); err != nil {
					log.Fatal(err.Error())
				}
			}
		}(conn)

	}
}

type HTTPrequest struct {
	Method    string
	Path      string
	Protocol  string
	UserAgent string
	Host      string
	Body      []byte
}

type HTTPresponseWriter struct {
	Protocol      string
	StatusCode    int
	StatusMessage string
	Headers       map[string]string
	Content       []byte
	w             io.Writer
}

func ParseRequest(content []byte) (*HTTPrequest, error) {
	r := HTTPrequest{}

	splits := bytes.Split(content, []byte("\r\n\r\n"))
	header := splits[0]
	r.Body = content[len(header)+4:]

	for index, line := range bytes.Split(header, []byte("\r\n")) {
		fmt.Printf("line %d: %s\n", index, string(line))
		if index == 0 {

			segments := strings.Split(string(line), " ")

			if len(segments) != 3 {
				return nil, errors.New("invalid header. ")
			}

			r.Method = segments[0]
			r.Path = segments[1]
			r.Protocol = segments[2]

		} else {

			segments := strings.Split(string(line), ": ")
			if len(segments) != 2 {
				return nil, errors.New("invalid header. ")
			}

			switch segments[0] {
			case "User-Agent":
				r.UserAgent = segments[1]
			case "Host":
				r.Host = segments[1]
			default:
				continue
			}
		}
	}
	return &r, nil
}

func (rw HTTPresponseWriter) Write() error {
	resp := []byte{}

	resp = append(resp, []byte(fmt.Sprintf("%s %d %s\r\n", rw.Protocol, rw.StatusCode, rw.StatusMessage))...)
	for key, value := range rw.Headers {
		resp = append(resp, []byte(fmt.Sprintf("%s: %s\r\n", key, value))...)
	}
	resp = append(resp, []byte("\r\n")...)
	resp = append(resp, rw.Content...)

	fmt.Println(string(resp))

	if _, err := rw.w.Write(resp); err != nil {
		return err
	}

	return nil
}

func CreateResponseWriter(w io.Writer, r *HTTPrequest) HTTPresponseWriter {
	return HTTPresponseWriter{
		Protocol: r.Protocol,
		Headers:  make(map[string]string),
		Content:  []byte{},
		w:        w,
	}
}

func HandleFunc(w HTTPresponseWriter, r *HTTPrequest) error {
	if r.Path == "/" {
		w.StatusCode = 200
		w.StatusMessage = "OK"
		return w.Write()
	}

	if r.Path == "/user-agent" {
		w.StatusCode = 200
		w.StatusMessage = "OK"
		w.Content = []byte(r.UserAgent)
		w.Headers["Content-Type"] = "text/plain"
		w.Headers["Content-Length"] = fmt.Sprintf("%d", len(w.Content))
		return w.Write()
	}

	if strings.HasPrefix(r.Path, "/echo/") {
		message := strings.TrimPrefix(r.Path, "/echo/")
		w.StatusCode = 200
		w.StatusMessage = "OK"
		w.Content = []byte(message)
		w.Headers["Content-Type"] = "text/plain"
		w.Headers["Content-Length"] = fmt.Sprintf("%d", len(w.Content))
		return w.Write()
	}

	if strings.HasPrefix(r.Path, "/files/") {
		return HandleFileFunc(w, r)
	}

	w.StatusCode = 404
	w.StatusMessage = "Not Found"
	return w.Write()
}

func HandleFileFunc(w HTTPresponseWriter, r *HTTPrequest) error {
	filename := strings.TrimPrefix(r.Path, "/files/")

	switch r.Method {
	case "GET":
		content, err := os.ReadFile(filename)

		if err != nil {
			w.StatusCode = 404
			w.StatusMessage = "Not Found"
			w.Headers["Content-Length"] = "0"

			return w.Write()
		}

		w.StatusCode = 200
		w.StatusMessage = "OK"
		w.Content = content
		w.Headers["Content-Type"] = "application/octet-stream"
		w.Headers["Content-Length"] = fmt.Sprintf("%d", len(w.Content))
		return w.Write()

	case "POST":
		content := r.Body
		err := os.WriteFile(filename, content, 0644)

		if err != nil {
			w.StatusCode = 500
			w.StatusMessage = "Internal Server Error"
			w.Headers["Content-Length"] = "0"

			return w.Write()
		}

		w.StatusCode = 201
		w.StatusMessage = "Created"
		w.Headers["Content-Length"] = "0"

		return w.Write()
	}

	w.StatusCode = 404
	w.StatusMessage = "Not Found"
	w.Headers["Content-Length"] = "0"

	return w.Write()
}
