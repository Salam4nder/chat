package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Salam4nder/chat/internal/config"
	"github.com/Salam4nder/chat/internal/db/cql/migrate"

	"github.com/gocql/gocql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestMain(m *testing.M) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	config := config.ScyllaDB{
		Hosts:             []string{"127.0.0.1"},
		Keyspaces:         []string{"message"},
		Namespace:         "chat",
		ReplicationFactor: 3,
	}

	waitForScylla(config.Hosts...)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := migrate.Run(
		ctx,
		config.Hosts,
		config.Namespace,
		config.ReplicationFactor,
		config.Keyspaces,
	); err != nil {
		log.Error().Err(err).Msg("failed to migrate")
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func waitForScylla(hosts ...string) {
	const (
		maxAttempts = 30
		sleepTime   = 1 * time.Second
	)

	for i := 0; i < maxAttempts; i++ {
		cluster := gocql.NewCluster(hosts...)
		cluster.Consistency = gocql.Quorum
		cluster.Keyspace = "system_schema"
		session, err := cluster.CreateSession()
		if err == nil {
			session.Close()
			return
		}

		log.Info().Msgf("waiting for ScyllaDB, attempt %d/%d:", i+1, maxAttempts)
		time.Sleep(sleepTime)
	}
}
