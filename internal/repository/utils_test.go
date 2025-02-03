package repository_test

import (
	"context"
	"crypto/rand"
	"io"
	"log"
	mrand "math/rand/v2"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	assert "github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/zanz1n/blog/internal/utils"
	"github.com/zanz1n/blog/migrations"
)

var lazyDb = utils.NewLazy(initDb)

func GetDb(t *testing.T) *sqlx.DB {
	db, err := lazyDb.Get()
	if err != nil {
		log.Printf("❌ Failed to init database: %s\n", err)
	}

	assert.NoError(t, err)
	return db
}

func initDb() (*sqlx.DB, error) {
	driver := "sqlite3"
	dialect := "sqlite3"
	mpath := "sqlite"
	url := "file::memory:?cache=shared"

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

	log.Printf("✅ Using %s for repository tests\n", dialect)

	db, err := sqlx.Open(driver, url)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Connected to %s instance\n", dialect)

	err = goose.SetDialect(dialect)
	if err != nil {
		return nil, err
	}

	goose.SetBaseFS(migrations.EmbedMigrations)
	goose.SetLogger(log.New(io.Discard, "", log.Flags()))

	err = goose.Up(db.DB, mpath)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Migrations completed on %s\n", dialect)

	log.Printf("✅ Sucessfully initialized %s\n", dialect)

	return db, err
}

func launchPostgresCt() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
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
				WithStartupTimeout(10*time.Second)),
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
		b[i] = letterRunes[mrand.IntN(len(letterRunes))]
	}
	return string(b)
}

func randBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}
