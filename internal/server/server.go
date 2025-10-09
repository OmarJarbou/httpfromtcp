package server

import (
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
		if s.Closed.Load() {
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
	writer := response.Writer{
		Writer:      conn,
		WriterState: response.STATUS_LINE,
	}
	req, err := request.RequestFromReader(conn)
	if err != nil {
		handler_response := &HandlerResponse{
			StatusCode: response.SERVER_ERROR,
			Message:    err.Error(),
		}
		handler_response.HandlerResponseWriter(writer)
	}

	s.Handler(writer, req)
}
