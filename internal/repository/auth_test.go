package repository_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const issuer = "SRV"

func authRepository(assert *require.Assertions) *repository.AuthRepository {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(err)

	return repository.NewAuthRepository(priv, pub, issuer)
}

func newUser(assert *require.Assertions) dto.User {
	user, err := dto.NewUser(userData(), dto.PermissionDefault, bcrypt.MinCost)
	assert.NoError(err)

	return user
}

func TestJwtEncodeDecode(t *testing.T) {
	assert := require.New(t)
	repo := authRepository(assert)

	user := newUser(assert)
	data := dto.NewAuthToken(&user, issuer, time.Hour)

	token, err := repo.EncodeToken(data)
	assert.NoError(err)

	data2, err := repo.DecodeToken(token)
	assert.NoError(err)

	assert.Equal(data, data2)
}

func TestJwtIssuerSet(t *testing.T) {
	assert := require.New(t)
	repo := authRepository(assert)

	user := newUser(assert)
	data := dto.NewAuthToken(&user, "", time.Hour)

	token, err := repo.EncodeToken(data)
	assert.NoError(err)

	data2, err := repo.DecodeToken(token)
	assert.NoError(err)

	data.Issuer = issuer
	assert.Equal(data, data2)
}

func TestJwtExpiration(t *testing.T) {
	assert := require.New(t)
	repo := authRepository(assert)

	user := newUser(assert)
	data := dto.NewAuthToken(&user, issuer, time.Second)

	token, err := repo.EncodeToken(data)
	assert.NoError(err)

	_, err = repo.DecodeToken(token)
	assert.NoError(err)

	time.Sleep(2 * time.Second)

	_, err = repo.DecodeToken(token)
	assert.Error(err)
	assert.ErrorIs(err, repository.ErrExpiredAuthToken)
}

func TestJwtParseInvalid(t *testing.T) {
	assert := require.New(t)
	repo1 := authRepository(assert)
	repo2 := authRepository(assert)

	user := newUser(assert)
	data := dto.NewAuthToken(&user, "", time.Hour)

	token, err := repo1.EncodeToken(data)
	assert.NoError(err)

	_, err = repo2.DecodeToken(token)
	assert.Error(err)
	assert.ErrorIs(err, repository.ErrInvalidAuthToken)

	_, err = repo2.DecodeToken("BLABLABLA")
	assert.Error(err)
	assert.ErrorIs(err, repository.ErrInvalidAuthToken)
}
