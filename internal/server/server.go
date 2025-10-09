package server

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/OmarJarbou/httpfromtcp/internal/request"
	"github.com/OmarJarbou/httpfromtcp/internal/response"
)

type Server struct {
	Listener net.Listener
	Handler  Handler
	Closed   atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	server := Server{}
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return &server, err
	}
	server.Handler = handler
	server.Listener = listener

	go server.listen()

	return &server, nil
}

func (s *Server) Close() error {
	err := s.Listener.Close()
	s.Closed.Store(true)
	return err
}

func (s *Server) listen() {
	for !s.Closed.Load() {
		connection, err := s.Listener.Accept()
		if !s.Closed.Load() {
			return
		}
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Println("A connection has been accepted")
		go s.handle(connection)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		handler_err := &HandlerError{
			StatusCode: response.SERVER_ERROR,
			Message:    err.Error(),
		}
		handler_err.handlerResponseWriter(conn, bytes.NewBuffer([]byte(handler_err.Message)))
	}

	handler_data := []byte{}
	buffer := bytes.NewBuffer(handler_data)
	handler_err := s.Handler(buffer, req)

	handler_err.handlerResponseWriter(conn, buffer)
}
