package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Salam4nder/chat/internal/chat"
	"github.com/Salam4nder/chat/internal/config"
	"github.com/Salam4nder/chat/internal/http/handler/health"
	"github.com/Salam4nder/chat/internal/http/handler/websocket"
	"github.com/gocql/gocql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	// HTTPReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	HTTPReadTimeout = 10 * time.Second
	// HTTPWriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read.
	HTTPWriteTimeout = 10 * time.Second
	// EnvironmentDev is the development environment.
	EnvironmentDev = "dev"
	// ScyllaTimeout is the maximum duration to wait for a ScyllaDB connection.
	ScyllaTimeout = 30 * time.Second
)

func main() {
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	config, err := config.New()
	exitOnError(err)
	go config.Watch()

	if config.Environment == EnvironmentDev {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	log.Info().Str("service", config.ServiceName).Send()

	scyllaSession, err := connectToScyllaWithTimeout(config.ScyllaDB, ScyllaTimeout, sigCh)
	exitOnError(err)

	chat.Rooms = make(map[string]*chat.Room)

	server := http.Server{
		Addr:         config.HTTPServer.Addr(),
		Handler:      nil,
		ReadTimeout:  HTTPReadTimeout,
		WriteTimeout: HTTPWriteTimeout,
	}
	healthHandler := health.NewHandler(scyllaSession)
	http.HandleFunc("/health", healthHandler.Health)
	http.HandleFunc("/chat", websocket.HandleWS)
	go func() {
		log.Info().
			Str("addr", config.HTTPServer.Addr()).
			Msg("main: serving http server...")

		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				exitOnError(err)
			}
		}
	}()

	<-sigCh
	log.Info().Msg("main: starting graceful shutdown...")

	scyllaSession.Close()

	if err := server.Shutdown(context.Background()); err != nil {
		log.Error().
			Err(err).
			Msg("main: failed to shutdown http server")
		os.Exit(1)
	}
	log.Info().Msg("main: cleanup finished")
	os.Exit(0)
}

func exitOnError(err error) {
	if err != nil {
		log.Error().
			Err(err).
			Msg("main: failed to start service")
		os.Exit(1)
	}
}

func connectToScyllaWithTimeout(
	config config.ScyllaDB,
	timeout time.Duration,
	cancelCh chan os.Signal,
) (*gocql.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cluster := gocql.NewCluster(config.Hosts...)
	cluster.Keyspace = config.Keyspaces[0]

	log.Info().Msg("main: trying to connect to ScyllaDB...")

	var (
		err     error
		session *gocql.Session
	)
	for {
		select {
		case <-time.After(1 * time.Second):
			session, err = cluster.CreateSession()
			if session != nil {
				return session, nil
			}
			log.Info().Msgf("main: failed attempt to connect ScyllaDB: %v, retrying", err)

		case <-ctx.Done():
			if err == nil {
				return nil, fmt.Errorf("failed to connect to ScyllaDB: %w", err)
			}

		case <-cancelCh:
			return nil, fmt.Errorf("failed to connect to ScyllaDB, cancel signal received")
		}
	}
}
