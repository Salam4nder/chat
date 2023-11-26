// Package migrate is a helper package that provides functions for creating keyspaces
// and registering callbacks for logging.
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

// Run runs migrations.
func Run(
	ctx context.Context,
	hosts []string,
	nameSpace string,
	repFactor int,
	keyspaces []string,
) error {
	if err := createKeyspaces(hosts, nameSpace, repFactor, keyspaces); err != nil {
		return err
	}

	cluster := gocql.NewCluster(hosts...)
	cluster.Consistency = gocql.Consistency(1)
	cluster.Keyspace = nameSpace

	session, err := gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		return fmt.Errorf("migrate: failed to create session: %w", err)
	}
	defer session.Close()

	if err := registerLoggingCallbacks(); err != nil {
		return err
	}

	if err := migrate.FromFS(ctx, session, cql.Files); err != nil {
		return fmt.Errorf("migrate: failed to run migrations: %w", err)
	}

	return nil
}

func createKeyspaces(hosts []string, nameSpace string, repFactor int, keyspaces []string) error {
	cluster := gocql.NewCluster(hosts...)
	cluster.Consistency = gocql.Consistency(1)

	rootSess, err := gocql.NewSession(*cluster)
	if err != nil {
		return fmt.Errorf("migrate: failed to create session: %w", err)
	}
	defer rootSess.Close()

	if err := rootSess.Query(
		fmt.Sprintf(
			`CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION =
            { 'class' : 'SimpleStrategy', 'replication_factor' : %d }`,
			nameSpace,
			repFactor,
		),
	).Exec(); err != nil {
		return fmt.Errorf("migrate: failed to create namespace: %w", err)
	}

	cluster.Keyspace = nameSpace
	session, err := gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		return fmt.Errorf("migrate: failed to create session: %w", err)
	}
	defer session.Close()

	for _, keyspace := range keyspaces {
		if err := session.ExecStmt(
			fmt.Sprintf(
				`CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION =
            { 'class' : 'SimpleStrategy', 'replication_factor' : %d }`,
				keyspace,
				repFactor,
			),
		); err != nil {
			return fmt.Errorf("migrate: failed to create keyspace(s): %w", err)
		}
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
		log.Info().Msgf("found migration file %s", name)
		return nil
	}

	afterLog := func(
		_ context.Context,
		_ gocqlx.Session,
		_ migrate.CallbackEvent,
		name string,
	) error {
		log.Info().Msgf("%s successfully migrated", name)
		return nil
	}

	filesNames, err := fs.Glob(cql.Files, "*.cql")
	if err != nil {
		return fmt.Errorf("migrate: failed to get files names: %w", err)
	}

	reg := migrate.CallbackRegister{}
	for _, fileName := range filesNames {
		reg.Add(migrate.BeforeMigration, fileName, beforeLog)
		reg.Add(migrate.AfterMigration, fileName, afterLog)
	}
	migrate.Callback = reg.Callback

	return nil
}
