package backend

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/spoolman-bambu-bridge/internal/bridge"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type wsClient struct {
	conn *websocket.Conn
	send chan []byte
}

// Hub manages WebSocket connections and broadcasts state updates.
type Hub struct {
	mu      sync.Mutex
	clients map[*wsClient]struct{}
	bridge  *bridge.Bridge
}

// NewHub creates a new WebSocket hub.
func NewHub(b *bridge.Bridge) *Hub {
	return &Hub{
		clients: make(map[*wsClient]struct{}),
		bridge:  b,
	}
}

// HandleWS upgrades an HTTP connection to WebSocket and registers it.
func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &wsClient{
		conn: conn,
		send: make(chan []byte, 16),
	}

	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()

	// Writer goroutine — sole owner of conn writes
	go func() {
		for msg := range client.send {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				break
			}
		}
		conn.Close()
	}()

	// Reader — any message from client triggers a state push
	defer func() {
		h.mu.Lock()
		delete(h.clients, client)
		h.mu.Unlock()
		close(client.send)
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		select {
		case client.send <- h.bridge.State.SnapshotJSON():
		default:
		}
	}
}

// Broadcast sends a message to all connected WebSocket clients.
func (h *Hub) Broadcast(data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		select {
		case client.send <- data:
		default:
			// slow client, drop message
		}
	}
}
