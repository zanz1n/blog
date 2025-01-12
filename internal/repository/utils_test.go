package repository_test

import (
	"context"
	"io"
	"log"
	"log/slog"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/zanz1n/blog/internal/utils"
	"github.com/zanz1n/blog/migrations"
)

var onceMigrate sync.Once
var lazyDb utils.Lazy[sqlx.DB]

func init() {
	lazyDb = utils.NewLazy(initDb)
}

func GetDb(t *testing.T) *sqlx.DB {
	var (
		db  *sqlx.DB
		err error
	)
	if testing.Short() {
		db, err = initDb()
		t.Cleanup(func() {
			db.Close()
		})
	} else {
		db, err = lazyDb.Get()
	}

	require.NoError(t, err)
	return db
}

func initDb() (*sqlx.DB, error) {
	driver := "sqlite3"
	dialect := "sqlite3"
	mpath := "sqlite"
	url := "file::memory:"

	if !testing.Short() {
		driver = "pgx/v5"
		dialect = "postgres"
		mpath = "postgres"
		endpoint, err := launchPostgresCt()
		if err != nil {
			return nil, err
		}
		url = endpoint
	}

	db, err := sqlx.Open(driver, url)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	onceMigrate.Do(func() {
		if err := goose.SetDialect(dialect); err != nil {
			panic(err)
		}
		goose.SetBaseFS(migrations.EmbedMigrations)
		goose.SetLogger(log.New(io.Discard, "", log.Flags()))
	})

	if err = goose.Up(db.DB, mpath); err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return db, err
}

func launchPostgresCt() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	container, err := postgres.Run(
		ctx,
		"postgres:17",
		postgres.WithDatabase(randString(10)),
		postgres.WithUsername(randString(10)),
		postgres.WithPassword(randString(128)),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)

	if err != nil {
		return "", err
	}

	return container.ConnectionString(ctx, "sslmode=disable")
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.IntN(len(letterRunes))]
	}
	return string(b)
}
