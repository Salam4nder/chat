package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Salam4nder/chat/internal/config"
	"github.com/Salam4nder/chat/internal/db/cql"
	"github.com/Salam4nder/chat/internal/db/migrate"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const timeout = 30 * time.Second

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	config, err := config.New()
	exitOnError(err)

	log.Info().Msg("migrate cmd: migration started")

	cluster := cql.NewClusterConfig(config.ScyllaDB)
	if err := cluster.PingWithTimeout(timeout, interrupt); err != nil {
		exitOnError(err)
	}

	if err := migrate.NewMigrator(cluster.Inner()).Run(
		context.TODO(),
		config.ScyllaDB.Keyspace,
		config.ScyllaDB.ReplicationFactor,
	); err != nil {
		exitOnError(err)
	}

	log.Info().Msg("migrate cmd: migration successful")
}

func exitOnError(err error) {
	if err != nil {
		log.Error().Err(err).Msg("migrate cmd: failed to migrate")
		os.Exit(1)
	}
}
