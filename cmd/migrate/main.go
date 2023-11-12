package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"github.com/Salam4nder/chat/internal/config"
	"github.com/Salam4nder/chat/internal/db/cql"
	"github.com/gocql/gocql"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/migrate"
)

func main() {
	config, err := config.New()
	exitOnError(err)

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msg("migration started")

	createKeyspaces(*config.ScyllaDB)

	cluster := gocql.NewCluster(config.ScyllaDB.Hosts...)
	cluster.Consistency = gocql.Consistency(1)
	cluster.Keyspace = config.ScyllaDB.Namespace

	session, err := gocqlx.WrapSession(cluster.CreateSession())
	exitOnError(err)
	defer session.Close()

	registerLoggingCallbacks()

	if err := migrate.FromFS(context.Background(), session, cql.Files); err != nil {
		exitOnError(err)
	}

	log.Info().Msg("migration successful")
}

func registerLoggingCallbacks() {
	beforeLog := func(_ context.Context, _ gocqlx.Session, _ migrate.CallbackEvent, name string) error {
		log.Info().Msgf("found migration file %s", name)
		return nil
	}
	afterLog := func(_ context.Context, _ gocqlx.Session, _ migrate.CallbackEvent, name string) error {
		log.Info().Msgf("%s successfully migrated", name)
		return nil
	}

	filesNames, err := fs.Glob(cql.Files, "*.cql")
	exitOnError(err)

	reg := migrate.CallbackRegister{}
	for _, fileName := range filesNames {
		reg.Add(migrate.BeforeMigration, fileName, beforeLog)
		reg.Add(migrate.AfterMigration, fileName, afterLog)
	}
	migrate.Callback = reg.Callback
}

func createKeyspaces(config config.ScyllaDB) {
	cluster := gocql.NewCluster(config.Hosts...)
	cluster.Consistency = gocql.Consistency(1)

	rootSess, err := gocql.NewSession(*cluster)
	exitOnError(err)
	defer rootSess.Close()

	if err := rootSess.Query(
		fmt.Sprintf(
			`CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION =
            { 'class' : 'SimpleStrategy', 'replication_factor' : %d }`,
			config.Namespace,
			config.ReplicationFactor,
		),
	).Exec(); err != nil {
		exitOnError(err)
	}

	cluster.Keyspace = config.Namespace
	session, err := gocqlx.WrapSession(cluster.CreateSession())
	exitOnError(err)
	defer session.Close()

	for _, keyspace := range config.Keyspaces {
		if err := session.ExecStmt(
			fmt.Sprintf(
				`CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION =
            { 'class' : 'SimpleStrategy', 'replication_factor' : %d }`,
				keyspace,
				config.ReplicationFactor,
			),
		); err != nil {
			exitOnError(err)
		}
	}
}

func exitOnError(err error) {
	if err != nil {
		log.Error().Err(err).Msg("migration: failed to migrate")
		os.Exit(1)
	}
}
