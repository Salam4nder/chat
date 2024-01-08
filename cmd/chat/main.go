package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Salam4nder/chat/internal/chat"
	"github.com/Salam4nder/chat/internal/config"
	"github.com/Salam4nder/chat/internal/db/cql"
	"github.com/Salam4nder/chat/internal/http/handler/health"
	"github.com/Salam4nder/chat/internal/http/handler/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	// scyllaTimeout is the maximum duration to wait for a ScyllaDB connection.
	scyllaTimeout = 30 * time.Second
	// httpReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	httpReadTimeout = 10 * time.Second
	// httpWriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read.
	httpWriteTimeout = 10 * time.Second
	// environmentDev is the development environment.
	environmentDev = "dev"
)

func main() {
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	config, err := config.New()
	exitOnError(err)
	if config.Environment == environmentDev {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	go config.Watch()

	log.Info().Str("service", config.ServiceName).Send()

	chat.Rooms = make(map[string]*chat.Room)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	cluster := cql.NewClusterConfig(config.ScyllaDB)
	if err := cluster.PingCluster(scyllaTimeout, sigCh); err != nil {
		exitOnError(err)
	}
	scyllaSession, err := cluster.Inner().CreateSession()
	if err != nil {
		exitOnError(err)
	}

	server := http.Server{
		Addr:         config.HTTPServer.Addr(),
		Handler:      nil,
		ReadTimeout:  httpReadTimeout,
		WriteTimeout: httpWriteTimeout,
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
