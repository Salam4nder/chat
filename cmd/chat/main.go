package main

import (
	"os"
	"time"

	"github.com/Salam4nder/chat/internal/chat"
	"github.com/Salam4nder/chat/internal/config"
	"github.com/Salam4nder/chat/internal/http"

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

	server := http.New().WithOptions(
		http.WithAddr(config.HTTPServer.Addr()),
		http.WithHandler(nil),
		http.WithTimeout(ReadTimeout, WriteTimeout),
	)
	http.InitRoutes()

	if err := server.Serve(); err != nil {
		exitOnError(err)
	}
}

func exitOnError(err error) {
	if err != nil {
		log.Error().Err(err).Msg("main: failed to start session")
		os.Exit(1)
	}
}