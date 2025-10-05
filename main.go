package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	file, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
		return
	}

	ch := getLinesChannel(file)
	for {
		line, ok := <-ch
		if !ok { // if channel closed
			os.Exit(0)
			return
		}
		fmt.Println("read: " + line)
	}
}
