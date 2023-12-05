package health

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gocql/gocql"
	"github.com/rs/zerolog/log"
)

const (
	StatusHealthy   = "Healthy"
	StatusUnhealthy = "Unhealthy"
	StatusStarting  = "Starting"
)

type Status struct {
	Health      string `json:"health"`
	Timestamp   string `json:"timestamp"`
	ServiceName string `json:"service_name"`
}

type Handler struct {
	status     Status
	scyllaConn *gocql.Session
}

func NewHandler(scyllaConn *gocql.Session) *Handler {
	return &Handler{
		status: Status{
			Health:      "Starting",
			Timestamp:   time.Now().Format(time.RFC3339),
			ServiceName: "chat",
		},
		scyllaConn: scyllaConn,
	}
}

func (x *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	status := x.check()

	w.Header().Set("Content-Type", "application/json")

	switch status.Health {
	case StatusHealthy:
		w.WriteHeader(http.StatusOK)
	case StatusStarting:
		w.WriteHeader(http.StatusServiceUnavailable)
	case StatusUnhealthy:
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Error().Err(err).Msg("health: error writing response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Check the health of the server.
// The different statuses are:
// - Starting (when the server is starting)
// - Healthy (when the server is ready to accept requests)
// - Unhealthy (when the server is not ready to accept requests).
func (x *Handler) check() Status {
	x.status.Health = StatusHealthy

	if x.scyllaConn == nil || x.scyllaConn.Closed() {
		x.status.Health = StatusUnhealthy
	}

	x.status.Timestamp = time.Now().Format(time.RFC3339)
	log.Info().
		Str("Health: %s", x.status.Health).
		Str("Timestamp: %s", x.status.Timestamp).
		Send()

	return x.status
}
