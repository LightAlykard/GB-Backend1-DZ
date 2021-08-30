package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	messages = make(chan string) // канал для бродкастинга сообщений из консоли сервера всем подключенным клиентам
	mu       sync.Mutex          // мьютекс для предотвращения гонки при доступе к списку клиентских соединений
)

func spammer(clients *[]net.Conn, ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case tickerTime := <-ticker.C:
			mu.Lock()
			for _, conn := range *clients {
				_, err := io.WriteString(conn, fmt.Sprintf("%s\n", tickerTime.Format(time.RFC850)))
				if err != nil {
					return
				}
			}
			mu.Unlock()
		case msg := <-messages:
			mu.Lock()
			for _, conn := range *clients {
				_, err := io.WriteString(conn, fmt.Sprintf("%s\n", msg))
				if err != nil {
					return
				}
			}
			mu.Unlock()
		}
	}
}

func handleConn(clients *[]net.Conn, ctx context.Context) {
	go spammer(clients, ctx)

	input := bufio.NewScanner(os.Stdin)
	fmt.Print("Type here (ctrl+C for exit) > \n")
	for input.Scan() {
		fmt.Print("Type here (ctrl+C for exit) > \n")
		messages <- input.Text()
	}
	if err := input.Err(); err != nil {
		fmt.Print("Bufio scanner err: %s", err)
	}

	select {
	case <-ctx.Done():
		mu.Lock()
		for _, conn := range *clients {
			if err := conn.Close(); err != nil {
				log.Fatal(err)
			}

		}
		mu.Unlock()
		return
	}
}

func main() {

	var (
		clients      = []net.Conn{}
		ctx, cancel  = context.WithCancel(context.Background())
		cancelSignal = make(chan os.Signal, 1)
		done         = make(chan bool, 1) // канал, при получении сообщения в который - return

		// обработка сигнала прерывания
		catchSignal = func(cancelFunc context.CancelFunc, l net.Listener) {
			sig := <-cancelSignal
			log.Println("Stop signal - %v", sig)
			done <- true
			cancelFunc()
			err := l.Close()
			if err != nil {
				log.Println(err)
			}
		}
	)

	signal.Notify(cancelSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	cfg := net.ListenConfig{
		KeepAlive: time.Minute,
	}
	l, err := cfg.Listen(ctx, "tcp", "127.127.127.127:9000")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("im started!")
	go catchSignal(cancel, l)

	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case d := <-done:
				log.Println("Server stopted ", d)
				return
			default:
				log.Println(err)
			}
		} else {
			mu.Lock()
			clients = append(clients, conn)
			mu.Unlock()
			log.Println("Get connection from client #%d", len(clients))
			go handleConn(&clients, ctx)
		}
	}

}
