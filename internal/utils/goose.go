package utils

import (
	"context"
	"io"
	"log"
	"log/slog"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/zanz1n/blog/migrations"
)

var migrateSetupOnce sync.Once

func MigrateUp(db *sqlx.DB) error {
	return MigrateUpContext(context.Background(), db, true)
}

func MigrateUpContext(ctx context.Context, db *sqlx.DB, logs bool) error {
	var (
		dir     string
		dialect string
		err     error
	)

	switch db.DriverName() {
	case "sqlite3", "sqlite":
		dir = "sqlite"
		dialect = "sqlite3"
	case "pgx", "postgres", "pgx/v5":
		dir = "postgres"
		dialect = "postgres"
	}

	migrateSetupOnce.Do(func() {
		err = gooseSetup(dialect, logs)
	})

	if err != nil {
		return err
	}

	return goose.UpContext(ctx, db.DB, dir)
}

func gooseSetup(dialect string, logs bool) (err error) {
	if err = goose.SetDialect(dialect); err != nil {
		return
	}

	goose.SetBaseFS(migrations.EmbedMigrations)

	if logs {
		logger := slog.NewLogLogger(slog.Default().Handler(), slog.LevelInfo)
		logger.SetPrefix("Database: ")

		goose.SetLogger(logger)
	} else {
		goose.SetLogger(log.New(io.Discard, "", log.Flags()))
	}

	return
}
