package http

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// Server ...
type Server struct {
	http   *http.Server
	health *health
}

type health struct {
	Status      string `json:"status"`
	Timestamp   string `json:"timestamp"`
	ServiceName string `json:"service_name"`
}

// Option is an option that configures an HTTP server.
type Option func(*Server)

// New returns a new HTTP server.
func New() *Server {
	return &Server{
		http: &http.Server{},
		health: &health{
			Status:    "Starting",
			Timestamp: time.Now().Format(time.RFC3339),
		},
	}
}

// Serve starts serving HTTP requests.
func (x *Server) Serve() error {
	log.Info().Msgf("Serving HTTP server on %s", x.http.Addr)

	x.Ping()

	return x.http.ListenAndServe()
}

// WithOptions configures the HTTP server with the provided options.
func (x *Server) WithOptions(opts ...Option) *Server {
	for _, opt := range opts {
		opt(x)
	}

	return x
}

// WithAddr configures the HTTP server with the provided address.
func WithAddr(addr string) Option {
	return func(x *Server) {
		x.http.Addr = addr
	}
}

// WithHandler configures the HTTP server with the provided handler.
func WithHandler(handler http.Handler) Option {
	return func(x *Server) {
		x.http.Handler = handler
	}
}

// WithTimeout configures the HTTP server with the provided read and write timeout.
func WithTimeout(read, write time.Duration) Option {
	return func(x *Server) {
		x.http.ReadTimeout = read
		x.http.WriteTimeout = write
	}
}

// WithServiceName configures the HTTP server with the provided service name.
func WithServiceName(name string) Option {
	return func(x *Server) {
		x.health.ServiceName = name
	}
}

// Ping checks the health of the server.
// The different statuses are:
// - Starting (when the server is starting)
// - Healthy (when the server is ready to accept requests)
// - Unhealthy (when the server is not ready to accept requests)
// - Stopping (when the server is shutting down)
// - Stopped (when the server is stopped)
func (x *Server) Ping() {
	x.health.Status = "Healthy"
	x.health.Timestamp = time.Now().Format(time.RFC3339)

	log.Info().Msgf("Health: %s", x.health.Status)
	log.Info().Msgf("Timestamp: %s", x.health.Timestamp)
}
