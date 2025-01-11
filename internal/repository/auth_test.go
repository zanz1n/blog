package repository_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const issuer = "SRV"

func authRepository(t *testing.T) *repository.AuthRepository {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(t, err)

	return repository.NewAuthRepository(priv, pub, issuer)
}

func newUser(t *testing.T) dto.User {
	user, err := dto.NewUser(userData(), dto.PermissionDefault, bcrypt.MinCost)
	assert.NoError(t, err)

	return user
}

func TestJwt(t *testing.T) {
	t.Parallel()
	repo := authRepository(t)

	user := newUser(t)
	data := dto.NewAuthToken(&user, issuer, time.Second)

	token, err := repo.EncodeToken(data)
	assert.NoError(t, err)

	t.Run("Decode", func(t *testing.T) {
		// t.Parallel()
		data2, err := repo.DecodeToken(token)
		assert.NoError(t, err)

		assert.Equal(t, data, data2)
	})

	t.Run("DecodeExpired", func(t *testing.T) {
		// t.Parallel()
		time.Sleep(2 * time.Second)

		_, err = repo.DecodeToken(token)
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrExpiredAuthToken)
	})

	t.Run("DecodeWrongKey", func(t *testing.T) {
		// t.Parallel()
		repo2 := authRepository(t)
		_, err = repo2.DecodeToken(token)
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrInvalidAuthToken)
	})

	t.Run("DecodeRandom", func(t *testing.T) {
		// t.Parallel()
		_, err = repo.DecodeToken("BLABLABLA")
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrInvalidAuthToken)
	})
}

func TestJwtIssuer(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	repo := authRepository(t)

	user := newUser(t)
	data := dto.NewAuthToken(&user, "", time.Hour)

	token, err := repo.EncodeToken(data)
	assert.NoError(err)

	data2, err := repo.DecodeToken(token)
	assert.NoError(err)

	data.Issuer = issuer
	assert.Equal(data, data2)
}
