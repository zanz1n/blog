package repository_test

import (
	"context"
	"fmt"
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
	assert "github.com/stretchr/testify/require"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/repository"
	"github.com/zanz1n/blog/migrations"
	"golang.org/x/crypto/bcrypt"
)

var onceGoose sync.Once

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(log.Writer(), &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo,
	})))
}

func inMemoryUserRepo(t *testing.T) *repository.UserRepository {
	db, err := sqlx.Open("sqlite3", "file::memory:")
	assert.NoError(t, err)

	userRepo, err := repository.NewUserRepository(db)
	assert.NoError(t, err)

	onceGoose.Do(func() {
		err = goose.SetDialect("sqlite3")
		assert.NoError(t, err)

		goose.SetBaseFS(migrations.EmbedMigrations)
		goose.SetLogger(log.New(io.Discard, "", log.Flags()))
	})

	err = goose.Up(db.DB, "sqlite")
	assert.NoError(t, err)

	t.Cleanup(func() {
		userRepo.Close()
		db.Close()
	})

	return userRepo
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.IntN(len(letterRunes))]
	}
	return string(b)
}

func userData() dto.UserCreateData {
	return dto.UserCreateData{
		Email:    fmt.Sprintf("%s@%s.com", randString(10), randString(20)),
		Nickname: randString(30),
		Name:     fmt.Sprintf("%s %s", randString(7), randString(10)),
		Password: randString(10),
	}
}

func TestUserCreate(t *testing.T) {
	t.Parallel()
	repo := inMemoryUserRepo(t)

	t.Run("Inexistent", func(t *testing.T) {
		// t.Parallel()
		_, err := repo.GetById(context.Background(), dto.NewSnowflake())
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrUserNotFound)
	})

	user, err := dto.NewUser(userData(), dto.PermissionDefault, bcrypt.MinCost)
	assert.NoError(t, err)

	err = repo.Create(context.Background(), user)
	assert.NoError(t, err)

	t.Run("Duplicate", func(t *testing.T) {
		// t.Parallel()
		err := repo.Create(context.Background(), user)
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrUserAlreadyExists)
	})

	t.Run("Fetch", func(t *testing.T) {
		// t.Parallel()
		user2, err := repo.GetById(context.Background(), user.ID)
		assert.NoError(t, err)
		assert.Equal(t, user, user2)

		user2, err = repo.GetByEmail(context.Background(), user.Email)
		assert.NoError(t, err)
		assert.Equal(t, user, user2)
	})
}

func TestUserUpdateName(t *testing.T) {
	t.Parallel()
	repo := inMemoryUserRepo(t)

	user, err := dto.NewUser(userData(), dto.PermissionDefault, bcrypt.MinCost)
	assert.NoError(t, err)

	err = repo.Create(context.Background(), user)
	assert.NoError(t, err)

	time.Sleep(5 * time.Millisecond)

	t.Run("Update", func(t *testing.T) {
		user2, err := repo.UpdateName(context.Background(), user.ID, "New Name")
		assert.NoError(t, err)

		user.Name = "New Name"
		assert.Greater(t, user2.UpdatedAt.UnixMilli(), user.UpdatedAt.UnixMilli())
		user.UpdatedAt = user2.UpdatedAt

		assert.Equal(t, user, user2)
	})

	t.Run("Fetch", func(t *testing.T) {
		user2, err := repo.GetById(context.Background(), user.ID)
		assert.NoError(t, err)
		assert.Equal(t, user, user2)
	})
}

func TestUserDelete(t *testing.T) {
	t.Parallel()
	repo := inMemoryUserRepo(t)

	t.Run("Inexistent", func(t *testing.T) {
		// t.Parallel()
		_, err := repo.DeleteById(context.Background(), dto.NewSnowflake())
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrUserNotFound)
	})

	user, err := dto.NewUser(userData(), dto.PermissionDefault, bcrypt.MinCost)
	assert.NoError(t, err)

	err = repo.Create(context.Background(), user)
	assert.NoError(t, err)

	t.Run("Delete", func(t *testing.T) {
		user2, err := repo.DeleteById(context.Background(), user.ID)
		assert.NoError(t, err)

		assert.Equal(t, user, user2)
	})

	t.Run("Fetch", func(t *testing.T) {
		_, err = repo.GetById(context.Background(), user.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrUserNotFound)
	})
}
