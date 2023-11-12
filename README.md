# chat
Chat application using ScyllaDB and Kafka.

## Migrations
The `cmd/migrate` tool provides a helper for CQL migrations, creating the necessary keyspaces as defined in the `config.yaml` file and executing migration files found in the `db/cql` directory.

Applied migrations will be logged to stdout. Do *not* edit applied migration files; instead, create new files for any edits. Migration files that have been successfully applied will be skipped in future migration runs.

To apply the migration, run either `make migrate` or `go run cmd/migration/main.go`.
