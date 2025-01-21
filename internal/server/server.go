package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"video-app-backend/internal/client"
	"video-app-backend/internal/models"
	"video-app-backend/internal/room"

	"github.com/gorilla/websocket"
)

// SignalingServer manages all rooms and connections
type SignalingServer struct {
	rooms    map[string]*room.Room
	upgrader websocket.Upgrader
	mu       sync.RWMutex
}

// NewSignalingServer creates a new signaling server instance
func NewSignalingServer() *SignalingServer {
	return &SignalingServer{
		rooms: make(map[string]*room.Room),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // In production, implement proper origin checking
			},
		},
	}
}

// getOrCreateRoom returns an existing room or creates a new one
func (s *SignalingServer) getOrCreateRoom(roomID string) *room.Room {
	s.mu.Lock()
	defer s.mu.Unlock()

	if room, exists := s.rooms[roomID]; exists {
		return room
	}

	newRoom := room.NewRoom(roomID)
	s.rooms[roomID] = newRoom
	return newRoom
}

// HandleWebSocket manages a WebSocket connection for a peer
func (s *SignalingServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	clientID := r.URL.Query().Get("clientId")
	if clientID == "" {
		log.Println("Client ID is required")
		return
	}

	c := client.NewClient(clientID, conn)

	for {
		var msg models.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			s.handleClientDisconnect(c)
			return
		}

		// Add debug logging
		log.Printf("Received WebSocket message from Frontend: Type=%s, Room=%s, From=%s, To=%s",
			msg.Type, msg.Room, msg.From, msg.Target)

		switch msg.Type {
		case "join":
			s.handleJoinRoom(c, &msg)
		case "leave":
			s.handleLeaveRoom(c)
		case "offer", "answer", "ice-candidate":
			s.handleSignaling(c, &msg)
		case "reconnect":
			s.handleReconnect(c, &msg)
		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

// handleJoinRoom processes a room join request
func (s *SignalingServer) handleJoinRoom(c *client.Client, msg *models.Message) {
	if msg.Room == "" {
		log.Println("Room ID is required")
		return
	}

	room := s.getOrCreateRoom(msg.Room)

	c.Room = msg.Room
	room.AddClient(c)

	// Get list of existing peers to send to the new client
	peers := room.GetClients()

	// Remove the new client from the peers list
	for i, peer := range peers {
		if peer == c.ID {
			peers = append(peers[:i], peers[i+1:]...)
			break
		}
	}

	// Notify the new peer about existing participants
	joinResponse := models.Message{
		Type:    "peers",
		From:    "server",
		Room:    msg.Room,
		Payload: json.RawMessage(`{"peers":` + marshal(peers) + `}`),
	}
	c.SendMessage(joinResponse)

	// Broadcast new peer arrival to all other peers in the room
	room.Broadcast(models.Message{
		Type: "peer-joined",
		From: c.ID,
		Room: msg.Room,
	}, c.ID)
}

// handleClientDisconnect processes a room unexpected disconnect
func (s *SignalingServer) handleClientDisconnect(c *client.Client) { // , msg *models.Message
	if c.Room == "" {
		return
	}

	s.mu.Lock()
	room, exists := s.rooms[c.Room]
	if !exists {
		s.mu.Unlock()
		return
	}

	room.RemoveClient(c.ID)

	// Remove room if empty
	if room.IsEmpty() {
		delete(s.rooms, c.Room)
	}
	s.mu.Unlock()
}

// handleSignaling routes WebRTC signaling messages between peers
func (s *SignalingServer) handleSignaling(c *client.Client, msg *models.Message) {
	if msg.Target == "" {
		log.Println("Target peer ID is required for signaling messages")
		return
	}

	s.mu.RLock()
	room, exists := s.rooms[msg.Room]
	s.mu.RUnlock()

	if !exists {
		log.Printf("Room %s not found", msg.Room)
		return
	}

	targetClient, exists := room.GetClient(msg.Target)
	if !exists {
		log.Printf("Target client %s not found in room %s", msg.Target, msg.Room)
		return
	}

	msg.From = c.ID
	targetClient.SendMessage(msg)
}

// handleReconnect processes a reconnect request
func (s *SignalingServer) handleReconnect(c *client.Client, msg *models.Message) {
	if msg.Room == "" {
		log.Println("Room ID is required")
		return
	}

	// Use getOrCreateRoom instead of just checking if room exists
	room := s.getOrCreateRoom(msg.Room)

	c.Room = msg.Room
	room.AddClient(c)

	// Get list of existing peers to send to the reconnecting client
	peers := room.GetClients()

	// Remove the reconnecting client from the peers list
	for i, peer := range peers {
		if peer == c.ID {
			peers = append(peers[:i], peers[i+1:]...)
			break
		}
	}

	// Notify the reconnecting client about existing participants
	reconnectResponse := models.Message{
		Type:    "reconnect-peers",
		From:    "server",
		Room:    msg.Room,
		Payload: json.RawMessage(`{"peers":` + marshal(peers) + `}`),
	}
	c.SendMessage(reconnectResponse)
}

// handleLeaveRoom cleans up when a client disconnects
func (s *SignalingServer) handleLeaveRoom(c *client.Client) {
	fmt.Println("Inside leave function")
	if c.Room == "" {
		return
	}

	s.mu.Lock()
	room, exists := s.rooms[c.Room]
	if !exists {
		s.mu.Unlock()
		return
	}

	room.RemoveClient(c.ID)

	// Remove room if empty
	if room.IsEmpty() {
		delete(s.rooms, c.Room)
	}
	s.mu.Unlock()

	// Broadcast departure to remaining peers
	if exists {
		room.Broadcast(models.Message{
			Type: "peer-left",
			From: c.ID,
			Room: c.Room,
		}, c.ID)
	}
}

// marshal is a helper function to convert data to JSON
func marshal(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
