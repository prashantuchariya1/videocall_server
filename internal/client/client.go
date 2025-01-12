package client

import (
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
	return c.Conn.WriteJSON(msg)
}
