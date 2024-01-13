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

var TestScyllaConn *ScyllaMessageRepository

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
	if err := cluster.PingCluster(timeout, interrupt); err != nil {
		exitOnError(err)
	}

	if err := migrate.NewMigrator(cluster.Inner()).Run(
		context.TODO(),
		config.Keyspace,
		config.ReplicationFactor,
	); err != nil {
		exitOnError(err)
	}

	session, err := cluster.Inner().CreateSession()
	exitOnError(err)

	TestScyllaConn = NewScyllaMessageRepository(session)

	os.Exit(m.Run())
}

func exitOnError(err error) {
	if err != nil {
		err = fmt.Errorf("message main_test: %w", err)
		log.Error().Err(err).Send()
		os.Exit(1)
	}
}
