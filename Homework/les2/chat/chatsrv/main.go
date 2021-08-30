package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

type client struct {
	Messages chan<- string
	Nickname string
}

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
)

func Server(address string, port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		log.Fatal(err)
	}

	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	timeout := time.NewTimer(100 * time.Second)
	ch := make(chan string)
	go clientWriter(conn, ch)

	enter := make(chan string) // входящие сообщения от клиентов
	go func() {
		input := bufio.NewScanner(conn)
		for input.Scan() {
			enter <- input.Text()
		}
	}()

	var who string
	ch <- "Enter your name: "
	who = <-enter

	cl := client{ch, who}
	messages <- "New user has arrived"
	entering <- cl

loop:
	for {
		select {
		case m := <-enter:
			messages <- who + " : " + m
		case <-timeout.C: // если пользователь при подключении ничего не вводит - то закрываем соединение
			err := conn.Close()
			if err != nil {
				log.Fatal(err)
			}
			break loop
		}

	}

	leaving <- cl
	messages <- who + " has left"
	err := conn.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		_, err := fmt.Fprintln(conn, msg)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli.Messages <- msg
			}
		case cli := <-entering:
			clients[cli] = true
			for clie := range clients {
				clie.Messages <- "Nick: " + cli.Nickname
			}

		case cli := <-leaving:
			delete(clients, cli)
			close(cli.Messages)
		}
	}
}

func main() {
	Server("localhost", 8000)
}
