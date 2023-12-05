//go:build testdb

package message

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/Salam4nder/chat/internal/config"
	"github.com/Salam4nder/chat/internal/db/migrate"
	"github.com/gocql/gocql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	waitTimeout    = 30 * time.Second
	migrateTimeout = 15 * time.Second
)

var TestScyllaConn *ScyllaKeyspace

func TestMain(m *testing.M) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	config := config.ScyllaDB{
		Hosts:             []string{"127.0.0.1"},
		Keyspaces:         []string{"message"},
		Namespace:         "chat",
		ReplicationFactor: 3,
	}

	waitCtx, waitCancel := context.WithTimeout(context.Background(), waitTimeout)
	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), migrateTimeout)

	if err := waitForScylla(waitCtx, sigCh, config.Hosts...); err != nil {
		log.Warn().Err(err).Send()
		waitCancel()
		os.Exit(1)
	}

	if err := migrate.Run(
		migrateCtx,
		config.Hosts,
		config.Namespace,
		config.ReplicationFactor,
		config.Keyspaces,
	); err != nil {
		log.Error().Err(err).Msg("repository: failed to migrate")
		migrateCancel()
		os.Exit(1)
	}

	cluster := gocql.NewCluster(config.Hosts...)
	cluster.Consistency = gocql.Consistency(1)
	cluster.Keyspace = config.Keyspaces[0]
	session, err := cluster.CreateSession()
	if err != nil {
		log.Error().Err(err).Msg("repository: failed to create session")
		os.Exit(1)
	}
	defer session.Close()

	TestScyllaConn = NewKeyspace(session)

	waitCancel()
	migrateCancel()
	os.Exit(m.Run())
}

func waitForScylla(ctx context.Context, sigCh <-chan os.Signal, hosts ...string) error {
	const (
		maxAttempts    = 30
		sleepTime      = 1 * time.Second
		systemKeyspace = "system_schema"
	)

	for i := 0; i < maxAttempts; i++ {
		select {
		case <-time.After(sleepTime):
			log.Info().Msgf("waiting for ScyllaDB, attempt %d/%d:", i+1, maxAttempts)
			cluster := gocql.NewCluster(hosts...)
			cluster.Consistency = gocql.Quorum
			cluster.Keyspace = systemKeyspace
			session, err := cluster.CreateSession()
			if err == nil {
				session.Close()
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		case <-sigCh:
			return errors.New("repository: waiting for scylla interrupted")
		}
	}

	return nil
}
