package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/valkey-io/valkey-go"
	"github.com/zanz1n/blog/config"
	"github.com/zanz1n/blog/internal/kv"
	"github.com/zanz1n/blog/internal/utils"
)

var migrateOpt = flag.Bool(
	"migrate",
	false,
	"executes database migrations before running the server",
)

func kvconnect(db *sqlx.DB) (kv.KVStorer, error) {
	cfg, err := config.Get()
	if err != nil {
		return nil, err
	}

	if cfg.RedisUrl == "" {
		slog.Info(
			fmt.Sprintf("KeyValue: Using %s instance", db.DriverName()),
		)
		return kv.NewSqlKV(db), nil
	}

	start := time.Now()

	url, err := valkey.ParseURL(cfg.RedisUrl)
	if err != nil {
		return nil, err
	}

	client, err := valkey.NewClient(url)
	if err != nil {
		return nil, err
	}

	repo := kv.NewRedisKV(client)

	slog.Info(
		"KeyValue: Connected to redis",
		utils.TookAttr(start, time.Microsecond),
	)

	return repo, nil
}

func dbconnect(ctx context.Context) (db *sqlx.DB, err error) {
	cfg, err := config.Get()
	if err != nil {
		return nil, err
	}

	start := time.Now()

	var scheme string
	if strings.HasPrefix(cfg.DatabaseUrl, "file:") {
		scheme = "sqlite"
		if err = touch(cfg.DatabaseUrl[len("file:"):]); err != nil {
			return
		}
		db, err = sqlx.Open("sqlite3", cfg.DatabaseUrl)
	} else {
		scheme = "postgres"
		db, err = sqlx.Open("pgx/v5", cfg.DatabaseUrl)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database pool: %s", err)
	}
	// if err = db.PingContext(ctx); err != nil {
	// 	db.Close()
	// 	return nil, fmt.Errorf("failed to ping database: %s", err)
	// }

	if *migrateOpt {
		if err = utils.MigrateUpContext(ctx, db, true); err != nil {
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
