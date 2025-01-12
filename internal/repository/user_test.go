package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

func userRepo(t *testing.T) *repository.UserRepository {
	db := GetDb(t)

	userRepo, err := repository.NewUserRepository(db)
	assert.NoError(t, err)

	return userRepo
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
	repo := userRepo(t)

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
	repo := userRepo(t)

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
	repo := userRepo(t)

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
