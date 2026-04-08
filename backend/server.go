package backend

import (
	"fmt"
	"net/http"

	"github.com/spoolman-bambu-bridge/frontend"
	"github.com/spoolman-bambu-bridge/internal/bridge"
)

// Version is set at build time via ldflags. Defaults to "dev".
var Version = "dev"

// Server holds the HTTP server and its dependencies.
type Server struct {
	mux    *http.ServeMux
	bridge *bridge.Bridge
	hub    *Hub
}

// NewServer creates a new backend server with API routes registered.
func NewServer(b *bridge.Bridge) *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		bridge: b,
		hub:    NewHub(b),
	}

	// Wire state changes to WebSocket broadcasts
	b.State.SetOnChange(func() {
		s.hub.Broadcast(b.State.SnapshotJSON())
	})

	s.registerRoutes()
	frontend.RegisterRoutes(s.mux)
	return s
}

// Handler returns the HTTP handler for this server.
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /api/v1/health", s.handleHealth)
	s.mux.HandleFunc("GET /ws", s.hub.HandleWS)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","version":"%s"}`, Version)
}
