package repository_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	mrand "math/rand/v2"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const issuer = "SRV"

func authRepository(t *testing.T, kv repository.KVStorer) *repository.AuthRepository {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(t, err)

	return repository.NewAuthRepository(priv, pub, issuer, kv)
}

func newUser(t *testing.T) dto.User {
	user, err := dto.NewUser(userData(), dto.PermissionDefault, bcrypt.MinCost)
	assert.NoError(t, err)

	return user
}

func TestAuthRefreshToken(t *testing.T) {
	t.Parallel()
	repo := authRepository(t, kvRepo(t))

	t.Run("Fuzz", func(t *testing.T) {
		t.Parallel()

		for i := range 32 {
			var token string
			if i%2 == 0 {
				token = randString(mrand.IntN(128))
			} else {
				token = base64.StdEncoding.EncodeToString(
					randBytes(repository.RefreshTokenLen),
				)
			}

			_, err := repo.ValidateRefreshToken(context.Background(), token)
			assert.Error(t, err)
			assert.ErrorIs(t, err, repository.ErrInvalidRefreshToken)
		}
	})

	userId := dto.NewSnowflake()

	token, err := repo.GenRefreshToken(context.Background(), userId)
	assert.NoError(t, err)

	_, err = repo.ValidateRefreshToken(context.Background(), token)
	assert.NoError(t, err)

	err = repo.DeleteRefreshTokens(context.Background(), userId)
	assert.NoError(t, err)

	_, err = repo.ValidateRefreshToken(context.Background(), token)
	assert.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrInvalidRefreshToken)
}

func TestAuthJwt(t *testing.T) {
	t.Parallel()
	repo := authRepository(t, nil)

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
		repo2 := authRepository(t, nil)
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

func TestAuthJwtIssuer(t *testing.T) {
	t.Parallel()
	repo := authRepository(t, nil)

	user := newUser(t)
	data := dto.NewAuthToken(&user, "", time.Hour)

	token, err := repo.EncodeToken(data)
	assert.NoError(t, err)

	data2, err := repo.DecodeToken(token)
	assert.NoError(t, err)

	data.Issuer = issuer
	assert.Equal(t, data, data2)
}
