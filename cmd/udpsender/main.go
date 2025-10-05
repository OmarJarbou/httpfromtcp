package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udp_addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal(err)
	}
	connection, err := net.DialUDP("udp", nil, udp_addr)
	// laddr: local UDP address (your computer) — can be nil to let the OS choose.
	// raddr: remote UDP address (the destination you’re sending to).
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		line, err := reader.ReadString(byte('\n')) // read up to newline
		if err != nil {
			log.Fatal(err)
		}
		_, err = connection.Write([]byte(line))
		if err != nil {
			log.Fatal(err)
		}
	}
}
