package main

import (
	"fmt"
	"log"
	"net"
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

		ch := getLinesChannel(connection)
		for {
			line, ok := <-ch
			if !ok { // if channel closed
				fmt.Println("Channel has been closed")
				break
			}
			fmt.Println(line)
		}
	}
}
