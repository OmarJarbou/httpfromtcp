package main

import (
	"fmt"
	"log"
	"net"

	"github.com/OmarJarbou/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("A connection has been accepted")

		req, err := request.RequestFromReader(connection)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Request line:")
		fmt.Println("- Method: " + req.RequestLine.Method)
		fmt.Println("- Target: " + req.RequestLine.RequestTarget)
		fmt.Println("- Version: " + req.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for key, value := range req.Headers {
			fmt.Println("- " + key + ": " + value)
		}

		fmt.Println("Body:")
		fmt.Println(string(req.Body))
	}
}
