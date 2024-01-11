package cql

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Salam4nder/chat/internal/config"
	"github.com/gocql/gocql"
	"github.com/rs/zerolog/log"
)

// ClusterConfig is a wrapper around gocql.ClusterConfig with helper methods
// like PingCluster.
type ClusterConfig struct {
	cluster *gocql.ClusterConfig
}

// NewClusterConfig creates a new ClusterConfig with the given configuration.
func NewClusterConfig(cfg config.ScyllaDB) *ClusterConfig {
	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Keyspace = cfg.Keyspace
	if cfg.Consistency == 4 {
		cluster.Consistency = gocql.Quorum
	} else {
		cluster.Consistency = gocql.Consistency(cfg.Consistency)
	}

	return &ClusterConfig{cluster: cluster}
}

// Inner returns the inner gocql.ClusterConfig.
func (x *ClusterConfig) Inner() *gocql.ClusterConfig {
	return x.cluster
}

// PingCluster creates a session with the 'system' keyspace and pings it.
// It returns an error if it cannot. Recommended timeout is 30 seconds.
// Resource cleanup is handled within the method.
func (x *ClusterConfig) PingCluster(
	timeout time.Duration,
	interrupt chan os.Signal,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var (
		err     error
		session *gocql.Session
	)

	keyspace := x.cluster.Keyspace
	x.cluster.Keyspace = "system"
	defer func() {
		x.cluster.Keyspace = keyspace
	}()

	for {
		select {
		case <-time.After(2 * time.Second):
			session, err = x.cluster.CreateSession()
			if session != nil {
				session.Close()
				return nil
			}
			log.Info().
				Err(err).
				Msg("ScyllaDB is not ready yet, retrying...")

		case <-ctx.Done():
			return fmt.Errorf("cql: pinging cluster: %w", ctx.Err())

		case <-interrupt:
			return errors.New("cql: interrupted")
		}
	}
}
