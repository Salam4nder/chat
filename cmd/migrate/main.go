package main

import (
	"context"
	"os"
	"time"

	"github.com/Salam4nder/chat/internal/config"
	"github.com/Salam4nder/chat/internal/db/cql/migrate"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	config, err := config.New()
	exitOnError(err)

	log.Info().Msg("migration started")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := migrate.Run(
		ctx,
		config.ScyllaDB.Hosts,
		config.ScyllaDB.Namespace,
		config.ScyllaDB.ReplicationFactor,
		config.ScyllaDB.Keyspaces,
	); err != nil {
		exitOnError(err)
	}

	log.Info().Msg("cmd: migration successful")
}

func exitOnError(err error) {
	if err != nil {
		log.Error().Err(err).Msg("cmd: failed to migrate")
		os.Exit(1)
	}
}
