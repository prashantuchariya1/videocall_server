package main

import (
	"log"
	"net/http"

	"video-app-backend/internal/server"
)

func main() {
	signalingServer := server.NewSignalingServer()

	// Set up HTTP routes
	http.HandleFunc("/ws", signalingServer.HandleWebSocket)

	// Start the server
	log.Println("Starting signaling server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
