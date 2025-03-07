package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/valkey-io/valkey-go"
	"github.com/zanz1n/blog/internal/repository"
	"github.com/zanz1n/blog/internal/utils"
	"github.com/zanz1n/blog/migrations"
)

var migrateOpt = flag.Bool(
	"migrate",
	false,
	"executes database migrations before running the server",
)

func kvconnect(db *sqlx.DB) (repository.KVStorer, error) {
	redisUrl := os.Getenv("REDIS_URL")

	if redisUrl == "" {
		slog.Info(
			fmt.Sprintf("KeyValue: Using %s instance", db.DriverName()),
		)
		return repository.NewSqlKV(db), nil
	}

	start := time.Now()

	url, err := valkey.ParseURL(redisUrl)
	if err != nil {
		return nil, err
	}

	client, err := valkey.NewClient(url)
	if err != nil {
		return nil, err
	}

	repo := repository.NewRedisKV(client)

	slog.Info(
		"KeyValue: Connected to redis",
		utils.TookAttr(start, time.Microsecond),
	)

	return repo, nil
}

func dbconnect() (db *sqlx.DB, err error) {
	dbUrl := os.Getenv("DATABASE_URL")

	start := time.Now()

	var scheme string
	if strings.HasPrefix(dbUrl, "file:") {
		scheme = "sqlite"
		if err = touch(dbUrl[len("file:"):]); err != nil {
			return
		}
		db, err = sqlx.Open("sqlite3", dbUrl)
	} else {
		scheme = "postgres"
		db, err = sqlx.Open("pgx/v5", dbUrl)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database pool: %s", err)
	}
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %s", err)
	}

	if *migrateOpt {
		if err = migrate(db.DB, scheme); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to run database migrations: %s", err)
		}
	}

	slog.Info(
		fmt.Sprintf("Database: Connected to %s", scheme),
		utils.TookAttr(start, time.Microsecond),
	)

	return
}

func migrate(db *sql.DB, dialect string) error {
	start := time.Now()
	if err := goose.SetDialect(dialect); err != nil {
		return err
	}

	logger := slog.NewLogLogger(slog.Default().Handler(), slog.LevelInfo)
	logger.SetPrefix("Database: ")

	goose.SetBaseFS(migrations.EmbedMigrations)
	goose.SetLogger(logger)

	err := goose.Up(db, dialect)
	if err != nil {
		return err
	}

	slog.Info(
		"Database: Executed migrations",
		utils.TookAttr(start, time.Microsecond),
	)
	return nil
}

func touch(name string) error {
	file, err := os.Open(name)
	if err != nil {
		file, err = os.Create(name)
	}
	if err != nil {
		return err
	}
	file.Close()
	return nil
}
