package client

import (
	"fmt"

	"video-app-backend/internal/models"

	"github.com/gorilla/websocket"
)

// Client represents a connected peer
type Client struct {
	ID   string
	Conn *websocket.Conn
	Room string
}

// NewClient creates a new client instance
func NewClient(id string, conn *websocket.Conn) *Client {
	return &Client{
		ID:   id,
		Conn: conn,
	}
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(msg interface{}) error {
	// Check if the message is of type Message
	if message, ok := msg.(models.Message); ok {
		// Log the required fields
		fmt.Printf(
			"Sending to client %s: {Type: %s, Target: %s, From: %s, Room: %s}\n",
			c.ID,
			message.Type,
			message.Target,
			message.From,
			message.Room,
		)
	} else {
		// Handle unexpected message format
		fmt.Printf("Sending to client %s: {Unsupported message format}\n", c.ID)
	}

	// Send the actual message as JSON
	return c.Conn.WriteJSON(msg)
}
