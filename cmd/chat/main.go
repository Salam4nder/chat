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
	"github.com/Salam4nder/chat/internal/http/handler/health"
	"github.com/Salam4nder/chat/internal/http/handler/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	ReadTimeout = 10 * time.Second
	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read.
	WriteTimeout = 10 * time.Second
	// EnvironmentDev is the development environment.
	EnvironmentDev = "dev"
)

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	config, err := config.New()
	exitOnError(err)
	go config.Watch()

	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if config.Environment == EnvironmentDev {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	log.Info().Str("service", config.ServiceName).Send()

	chat.Rooms = make(map[string]*chat.Room)

	server := http.Server{
		Addr:         config.HTTPServer.Addr(),
		Handler:      nil,
		ReadTimeout:  ReadTimeout,
		WriteTimeout: WriteTimeout,
	}

	healthHandler := health.NewHandler()

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

	if err := server.Shutdown(context.Background()); err != nil {
		log.Error().Err(err).Msg("main: failed to shutdown http server")
		os.Exit(1)
	}

	log.Info().Msg("main: cleanup finished")

	os.Exit(0)
}

func exitOnError(err error) {
	if err != nil {
		log.Error().Err(err).Msg("main: failed to start service")
		os.Exit(1)
	}
}
