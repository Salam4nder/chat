//go:build testdb

package chat

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/Salam4nder/chat/internal/config"
	"github.com/Salam4nder/chat/internal/db/cql"
	"github.com/Salam4nder/chat/internal/db/migrate"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	timeout  = 30 * time.Second
	keyspace = "chat"
)

var (
	testMessageRepo *ScyllaMessageRepository
	testUserRepo    *ScyllaUserRepository
)

func TestMain(m *testing.M) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	config := config.ScyllaDB{
		Hosts:             []string{"127.0.0.1"},
		Keyspace:          keyspace,
		ReplicationFactor: 3,
		Consistency:       1,
	}

	cluster := cql.NewClusterConfig(config)
	err := cluster.PingCluster(timeout, interrupt)
	exitOnError(err)

	err = migrate.NewMigrator(cluster.Inner()).Run(
		context.TODO(),
		config.Keyspace,
		config.ReplicationFactor,
	)
	exitOnError(err)

	session, err := cluster.Inner().CreateSession()
	exitOnError(err)

	testMessageRepo = NewScyllaMessageRepository(session)
	testUserRepo = NewScyllaUserRepository(session)

	os.Exit(m.Run())
}

func exitOnError(err error) {
	if err != nil {
		err = fmt.Errorf("message main_test: %w", err)
		log.Error().Err(err).Send()
		os.Exit(1)
	}
}
