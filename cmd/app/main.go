package main

import (
	"fmt"
	"net/http"

	"github.com/olahol/melody"
)

func main() {
	m := melody.New()

	// Event handler for new connections
	m.HandleConnect(func(session *melody.Session) {
		fmt.Println("New connection established")
	})

	// Event handler for disconnections
	m.HandleDisconnect(func(session *melody.Session) {
		fmt.Println("Connection closed")
	})

	// Event handler for received messages
	m.HandleMessage(func(session *melody.Session, message []byte) {
		// Broadcast the received message to all connected clients
		// m.Broadcast(message)
	})

	// Serve the chat application
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "chat.html")
	})

	// Upgrade HTTP requests to WebSocket connections
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		m.HandleRequest(w, r)
	})

	// Start the HTTP server
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
