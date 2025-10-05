package main

import (
	"io"
	"log"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)
	eight_bytes := make([]byte, 8)
	line := ""
	go func() {
		defer f.Close()
		defer close(ch)
		for {
			n, err := f.Read(eight_bytes)
			if n > 0 {
				line += string(eight_bytes[:n])
				split_by_new_line := strings.Split(line, "\n")
				if len(split_by_new_line) == 2 {
					ch <- split_by_new_line[0]
					line = split_by_new_line[1]
				}
			}
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Fatal(err)
				os.Exit(0)
				return
			}
		}
	}()

	return ch
}
