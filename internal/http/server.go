package http

import (
	"net/http"
	"time"

	"github.com/Salam4nder/chat/internal/config"

	"github.com/rs/zerolog"
)

// Server ...
type Server struct {
	http   *http.Server
	config *config.App
	logger *zerolog.Logger
	health *health
}

type health struct {
	Status      string `json:"status"`
	Timestamp   string `json:"timestamp"`
	ServiceName string `json:"service_name"`
}

// New returns a new HTTP server.
func New(
	cfg *config.App,
	log *zerolog.Logger,
) *Server {
	return &Server{
		config: cfg,
		logger: log,
		health: &health{
			Status:      "Starting",
			Timestamp:   time.Now().Format(time.RFC3339),
			ServiceName: "chat",
		},
	}
}
