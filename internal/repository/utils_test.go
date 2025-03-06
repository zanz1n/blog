package repository_test

import (
	"context"
	"io"
	"log"
	mrand "math/rand/v2"
	"sync"
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

var migrateOnce sync.Once

var lazyDb = utils.NewLazy(func() (*sqlx.DB, error) {
	return InitDb(nil)
})

func GetDb(t *testing.T) *sqlx.DB {
	var (
		db  *sqlx.DB
		err error
	)
	if !testing.Short() {
		db, err = lazyDb.Get()
	} else {
		db, err = InitDb(t)
		db.SetMaxOpenConns(1)
		t.Cleanup(func() {
			db.Close()
		})
	}

	if err != nil {
		log.Printf("❌ Failed to init database: %s\n", err)
	}

	assert.NoError(t, err)
	return db
}

func InitDb(t *testing.T) (*sqlx.DB, error) {
	short := testing.Short()

	driver := "sqlite3"
	dialect := "sqlite3"
	mpath := "sqlite"
	url := "file::memory:"

	if !short {
		driver = "pgx/v5"
		dialect = "postgres"
		mpath = "postgres"

		endpoint, err := launchPostgresCt()
		if err != nil {
			return nil, err
		}

		url = endpoint
	}

	if !short {
		log.Printf("✅ Using %s for repository tests\n", dialect)
	}

	db, err := sqlx.Open(driver, url)
	if err != nil {
		return nil, err
	}

	if !short {
		log.Printf("✅ Connected to %s instance\n", dialect)
	}

	gooseSetup(dialect)

	err = goose.Up(db.DB, mpath)
	if err != nil {
		return nil, err
	}

	if !short {
		log.Printf("✅ Migrations completed on %s\n", dialect)
		log.Printf("✅ Sucessfully initialized %s\n", dialect)
	}

	return db, err
}

func gooseSetup(dialect string) {
	migrateOnce.Do(func() {
		err := goose.SetDialect(dialect)
		if err != nil {
			panic(err)
		}

		goose.SetBaseFS(migrations.EmbedMigrations)
		goose.SetLogger(log.New(io.Discard, "", log.Flags()))
	})
}

func launchPostgresCt() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	container, err := postgres.Run(
		ctx,
		"postgres:17-alpine",
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
