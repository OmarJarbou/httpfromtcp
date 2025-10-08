package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/OmarJarbou/httpfromtcp/internal/response"
)

type Server struct {
	Listener net.Listener
	Closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	server := Server{}
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return &server, err
	}
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
		if err != nil && !s.Closed.Load() {
			log.Fatal(err)
		}
		fmt.Println("A connection has been accepted")
		go s.handle(connection)
	}
}

func (s *Server) handle(conn net.Conn) {
	err := response.WriteStatusLine(conn, response.OK)
	if err != nil {
		log.Fatal(err)
		return
	}
	headers := response.GetDefaultHeaders(0)
	err = response.WriteHeaders(conn, headers)
	if err != nil {
		log.Fatal(err)
		return
	}
}
