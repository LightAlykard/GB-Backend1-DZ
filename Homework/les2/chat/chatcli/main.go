package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
)

func Client(address string, port int) {
	conn, err := net.Dial("tcp", net.JoinHostPort(address, strconv.Itoa(port)))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Connected to the server %s:%d\n", address, port)
	defer conn.Close()
	go func() {
		_, err := io.Copy(os.Stdout, conn)
		if err != nil {
			log.Fatal(err)
		}
	}()
	_, err = io.Copy(conn, os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s: exit", conn.LocalAddr())
}

func main() {
	Client("localhost", 8000)
}
