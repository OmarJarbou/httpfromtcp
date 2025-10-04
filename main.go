package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	file, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatal(err)
		return
	}

	eight_bytes := make([]byte, 8)
	for {
		n, err := file.Read(eight_bytes)
		if n > 0 {
			fmt.Println("read:", string(eight_bytes[:n]))
		}
		if err == io.EOF {
			os.Exit(0)
			return
		}
		if err != nil {
			log.Fatal(err)
			return
		}
	}
}
