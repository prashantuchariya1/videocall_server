package room

import (
	"sync"

	"video-app-backend/internal/client"
	"video-app-backend/internal/models"
)

// Room manages a collection of peers in a video call
type Room struct {
	ID      string
	clients map[string]*client.Client
	mu      sync.RWMutex
}

// NewRoom creates a new room instance
func NewRoom(id string) *Room {
	return &Room{
		ID:      id,
		clients: make(map[string]*client.Client),
	}
}

// AddClient adds a client to the room thread-safely
func (r *Room) AddClient(client *client.Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[client.ID] = client
}

// RemoveClient removes a client from the room thread-safely
func (r *Room) RemoveClient(clientID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, clientID)
}

// GetClients returns a list of all client IDs in the room
func (r *Room) GetClients() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := make([]string, 0, len(r.clients))
	for clientID := range r.clients {
		clients = append(clients, clientID)
	}
	return clients
}

// GetClient returns a specific client by ID
func (r *Room) GetClient(clientID string) (*client.Client, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	client, exists := r.clients[clientID]
	return client, exists
}

// IsEmpty checks if the room has no clients
func (r *Room) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients) == 0
}

// Broadcast sends a message to all clients except one
func (r *Room) Broadcast(msg models.Message, exceptClientID string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for clientID, client := range r.clients {
		if clientID != exceptClientID {
			client.SendMessage(msg)
		}
	}
}
