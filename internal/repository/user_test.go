package repository_test

import (
	"context"
	"io"
	"log"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/repository"
	"github.com/zanz1n/blog/migrations"
	"golang.org/x/crypto/bcrypt"
)

func inMemoryUserRepo(t *testing.T) (*repository.UserRepository, *sqlx.DB) {
	assert := require.New(t)

	db, err := sqlx.Open("sqlite3", "file::memory:")
	assert.NoError(err)
	userRepo, err := repository.NewUserRepository(db)
	assert.NoError(err)

	err = goose.SetDialect("sqlite3")
	assert.NoError(err)

	goose.SetBaseFS(migrations.EmbedMigrations)
	goose.SetLogger(log.New(io.Discard, "", log.Flags()))

	err = goose.Up(db.DB, "sqlite")
	assert.NoError(err)

	return userRepo, db
}

func userData() dto.UserCreateData {
	return dto.UserCreateData{
		Email:    "johhdoe@example.com",
		Nickname: "johhdoe",
		Name:     "John Doe",
		Password: "strongpassword",
	}
}

func TestUserCreate(t *testing.T) {
	assert := require.New(t)

	repo, db := inMemoryUserRepo(t)
	defer db.Close()
	defer repo.Close()

	_, err := repo.GetById(context.Background(), dto.NewSnowflake())
	assert.Error(err)
	assert.ErrorIs(err, repository.ErrUserNotFound)

	data := userData()

	user, err := dto.NewUser(data, dto.PermissionDefault, bcrypt.MinCost)
	assert.NoError(err)

	err = repo.Create(context.Background(), user)
	assert.NoError(err)

	user2, err := repo.GetById(context.Background(), user.ID)
	assert.NoError(err)
	assert.Equal(user, user2)

	user2, err = repo.GetByEmail(context.Background(), user.Email)
	assert.NoError(err)
	assert.Equal(user, user2)
}

func TestUserUpdateName(t *testing.T) {
	assert := require.New(t)

	repo, db := inMemoryUserRepo(t)
	defer db.Close()
	defer repo.Close()

	user, err := dto.NewUser(userData(), dto.PermissionDefault, bcrypt.MinCost)
	assert.NoError(err)

	err = repo.Create(context.Background(), user)
	assert.NoError(err)

	time.Sleep(5 * time.Millisecond)

	user2, err := repo.UpdateName(context.Background(), user.ID, "New Name")
	assert.NoError(err)

	user.Name = "New Name"
	assert.Greater(user2.UpdatedAt.UnixMilli(), user.UpdatedAt.UnixMilli())
	user.UpdatedAt = user2.UpdatedAt

	assert.Equal(user, user2)

	user2, err = repo.GetById(context.Background(), user.ID)
	assert.NoError(err)
	assert.Equal(user, user2)
}

func TestUserDelete(t *testing.T) {
	assert := require.New(t)

	repo, db := inMemoryUserRepo(t)
	defer db.Close()
	defer repo.Close()

	_, err := repo.DeleteById(context.Background(), dto.NewSnowflake())
	assert.Error(err)
	assert.ErrorIs(err, repository.ErrUserNotFound)

	data := userData()

	user, err := dto.NewUser(data, dto.PermissionDefault, bcrypt.MinCost)
	assert.NoError(err)

	err = repo.Create(context.Background(), user)
	assert.NoError(err)

	user2, err := repo.DeleteById(context.Background(), user.ID)
	assert.NoError(err)

	assert.Equal(user, user2)

	_, err = repo.GetById(context.Background(), user.ID)
	assert.Error(err)
	assert.ErrorIs(err, repository.ErrUserNotFound)
}
