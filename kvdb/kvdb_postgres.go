//go:build kvdb_postgres
// +build kvdb_postgres

package kvdb

import "github.com/flokiorg/flnd/kvdb/postgres"

const PostgresBackend = true

func NewPostgresFixture(dbName string) (postgres.Fixture, error) {
	return postgres.NewFixture(dbName)
}

func StartEmbeddedPostgres() (func() error, error) {
	return postgres.StartEmbeddedPostgres()
}
