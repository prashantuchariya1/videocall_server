package models

import "encoding/json"

// Message represents a WebRTC signaling message
type Message struct {
	Type    string          `json:"type"`              // "offer", "answer", "ice-candidate", "join", "leave","reconnect"
	Target  string          `json:"target,omitempty"`  // Target peer ID (empty for broadcast)
	From    string          `json:"from"`              // Sender's peer ID
	Room    string          `json:"room"`              // Room identifier
	Payload json.RawMessage `json:"payload,omitempty"` // Actual message content (SDP or ICE candidate)
}
