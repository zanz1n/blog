package dto_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zanz1n/blog/internal/dto"
	"golang.org/x/crypto/bcrypt"
)

func userData() dto.UserCreateData {
	return dto.UserCreateData{
		Email:    "johhdoe@example.com",
		Nickname: "johhdoe",
		Name:     "John Doe",
		Password: "strongpassword",
	}
}

func TestUserPasswordMatches(t *testing.T) {
	assert := require.New(t)

	data := userData()
	user, err := dto.NewUser(data, dto.PermissionDefault, bcrypt.MinCost)
	assert.NoError(err)

	matches := user.PasswordMatches(data.Password)
	assert.True(matches)
}

func TestUserSnowflake(t *testing.T) {
	assert := require.New(t)

	user, err := dto.NewUser(userData(), dto.PermissionDefault, bcrypt.MinCost)
	assert.NoError(err)

	assert.Equal(user.ID.Timestamp(), user.CreatedAt.Time)
}
