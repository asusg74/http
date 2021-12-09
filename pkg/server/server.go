package server

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
	"sync"
)

var ErrNotSuatable = errors.New("string is not suatable for patter")

type HandlerFunc func(*Request)

type Server struct {
	addr     string
	mu       sync.RWMutex
	handlers map[string]HandlerFunc
}

type Request struct {
	Conn        net.Conn
	QueryParams url.Values
	PathParams  map[string]string
	Headers     map[string]string
	Pattern     string
	Body        []byte
}

func NewServer(addr string) *Server {
	return &Server{addr: addr, handlers: make(map[string]HandlerFunc)}
}

func (s *Server) Register(path string, handler HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[path] = handler
}

func CreateRequestFromPattern(path string, pattern string) (pathParams map[string]string, err error) {
	res := map[string]string{}
	j := 0
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '{' {
			key := ""
			val := ""
			for i++; i < len(pattern) && pattern[i] != '}'; {
				key += string(pattern[i])
				i++
			}
			for j < len(path) && path[j] != '/' {
				val += string(path[j])
				j++
			}
			res[key] = val
		} else if pattern[i] != path[j] {
			return nil, ErrNotSuatable
		} else {
			j++
		}
	}
	return res, nil
}

func handle(conn net.Conn, srv *Server) {
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()
	request := Request{Conn: conn, Headers: map[string]string{}}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err == io.EOF {
		buf = buf[:n]
	}
	if err != nil {
		log.Print(err)
	}
	data := buf[:n]
	requestLineDelim := []byte{'\r', '\n'}
	requestLineEnd := bytes.Index(data, requestLineDelim)
	headersLineEnd := requestLineEnd + bytes.Index(data[requestLineEnd+2:], []byte{'\r', '\n', '\r', '\n'})
	if requestLineEnd == -1 {
		log.Print("request headers not found")
		return
	}
	requestLine := string(data[:requestLineEnd])
	headersLine := string(data[requestLineEnd+2 : headersLineEnd+2])
	request.Body = data[headersLineEnd+6:]
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		log.Print("could not separate request")
		return
	}

	_, path, version := parts[0], parts[1], parts[2]

	for _, line := range strings.Split(strings.TrimSuffix(headersLine, "\r\n"), "\r\n") {
		if len(line) != 0 {
			pair := strings.SplitN(line, ": ", 2)
			request.Headers[pair[0]] = pair[1]
		}
	}

	if version != "HTTP/1.1" {
		log.Print("wrong http version")
		return
	}

	decoded, err := url.ParseRequestURI(path)
	path = decoded.Path
	request.QueryParams = decoded.Query()
	if err != nil {
		log.Print(err)
		return
	}
	exist := 0
	srv.mu.RLock()
	for k := range srv.handlers {
		if pathParams, err := CreateRequestFromPattern(path, k); err == nil {
			exist = 1
			request.PathParams = pathParams
			request.Pattern = k
			break
		}
	}
	srv.mu.RUnlock()

	if exist == 1 {
		srv.handlers[request.Pattern](&request)
	} else {
		log.Print("handler not found. Address: " + path)
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Print(err)
		return err
	}
	defer func() {
		if cerr := listener.Close(); cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(cerr)
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		go handle(conn, s)
	}
}
