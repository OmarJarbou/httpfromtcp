package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatal(err)
		return
	}

	eight_bytes := make([]byte, 8)
	line := ""
	for {
		n, err := file.Read(eight_bytes)
		if n > 0 {
			line += string(eight_bytes[:n])
			split_by_new_line := strings.Split(line, "\n")
			if len(split_by_new_line) == 2 {
				fmt.Println("read:", split_by_new_line[0])
				line = split_by_new_line[1]
			}
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
