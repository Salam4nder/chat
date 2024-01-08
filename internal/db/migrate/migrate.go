// Package migrate is a helper package that provides CQL migrations.
package migrate

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/Salam4nder/chat/internal/db/cql"
	"github.com/gocql/gocql"
	"github.com/rs/zerolog/log"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/migrate"
)

type Migrator struct {
	cluster *gocql.ClusterConfig
}

func NewMigrator(cluster *gocql.ClusterConfig) *Migrator {
	return &Migrator{cluster: cluster}
}

//Run runs migrations.
func (x *Migrator) Run(ctx context.Context, keyspace string, repFactor int) error {
	x.cluster.Keyspace = "system"
	cqlSession, err := x.cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("migrate: creating session: %w", err)
	}
	defer cqlSession.Close()

	if err := cqlSession.Query(
		fmt.Sprintf(
			`CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION =
            { 'class' : 'SimpleStrategy', 'replication_factor' : %d }`,
			keyspace,
			repFactor,
		),
	).Exec(); err != nil {
		return fmt.Errorf("migrate: creating keyspace: %w", err)
	}

	x.cluster.Keyspace = keyspace
	cqlxSession, err := gocqlx.WrapSession(x.cluster.CreateSession())
	if err != nil {
		return fmt.Errorf("migrate: wrapping session: %w", err)
	}
	defer cqlxSession.Close()

	if err := registerLoggingCallbacks(); err != nil {
		return err
	}

	if err := migrate.FromFS(ctx, cqlxSession, cql.Files); err != nil {
		return fmt.Errorf("migrate: migrating: %w", err)
	}

	return nil
}

func registerLoggingCallbacks() error {
	beforeLog := func(
		_ context.Context,
		_ gocqlx.Session,
		_ migrate.CallbackEvent,
		name string,
	) error {
		log.Info().Msgf("migrate: found migration file %s", name)
		return nil
	}

	afterLog := func(
		_ context.Context,
		_ gocqlx.Session,
		_ migrate.CallbackEvent,
		name string,
	) error {
		log.Info().Msgf("migrate: %s successfully migrated", name)
		return nil
	}

	filesNames, err := fs.Glob(cql.Files, "*.cql")
	if err != nil {
		return fmt.Errorf("migrate: getting files names: %w", err)
	}

	reg := migrate.CallbackRegister{}
	for _, fileName := range filesNames {
		reg.Add(migrate.BeforeMigration, fileName, beforeLog)
		reg.Add(migrate.AfterMigration, fileName, afterLog)
	}
	migrate.Callback = reg.Callback

	return nil
}
