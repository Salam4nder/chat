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
	internalHTTP "github.com/Salam4nder/chat/internal/http"
	"github.com/gocql/gocql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/scylladb/gocqlx/v2"
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

	cluster := gocql.NewCluster(config.ScyllaDB.Hosts...)

	session, err := gocqlx.WrapSession(cluster.CreateSession())
	exitOnError(err)
	defer session.Close()

	chat.Rooms = make(map[string]*chat.Room)

	httpServer := internalHTTP.New().WithOptions(
		internalHTTP.WithAddr(config.HTTPServer.Addr()),
		internalHTTP.WithHandler(nil),
		internalHTTP.WithTimeout(ReadTimeout, WriteTimeout),
	)
	internalHTTP.InitRoutes()

	go func() {
		if err := httpServer.Serve(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				exitOnError(err)
			}
		}
	}()

	<-sigCh
	log.Info().Msg("main: starting shutdown...")

	if err := httpServer.GracefulShutdown(context.Background()); err != nil {
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
