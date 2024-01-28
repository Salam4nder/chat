package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
)

var (
	addr         = flag.String("addr", "localhost:8080", "http service address")
	roomID       = flag.String("roomID", "", "room ID")
	friendlyName = flag.String("name", "", "friendly name")
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	u := url.URL{
		Scheme:   "ws",
		Host:     *addr,
		Path:     "/chat",
		RawQuery: fmt.Sprintf("%s=%s&%s=%s", "roomID", *roomID, "name", *friendlyName),
	}
	log.Printf("connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
		}
	}()

	inputStr := make(chan string)

	go func() {
		for {
			select {
			case t := <-inputStr:
				err := conn.WriteMessage(websocket.TextMessage, []byte(t))
				if err != nil {
					log.Println("write:", err)
					return
				}
			case <-interrupt:
				log.Println("interrupt")
				err := conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				)
				if err != nil {
					log.Println("write close:", err)
					os.Exit(1)
				}
				os.Exit(0)
			}
		}
	}()

	fmt.Println("Enter ':quit' to exit")

	for {
		buffer := make([]byte, 1024)

		n, err := os.Stdin.Read(buffer)
		if err != nil {
			fmt.Println(err)
		}

		if string(buffer[:n]) == ":quit\n" {
			interrupt <- os.Interrupt
			break
		}

		inputStr <- string(buffer[:n])
	}
}
