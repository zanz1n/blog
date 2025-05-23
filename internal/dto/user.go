package dto

import (
	"errors"
	"log/slog"
	"time"

	"github.com/zanz1n/blog/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID         Snowflake  `db:"id" json:"id"`
	CreatedAt  Timestamp  `db:"created_at" json:"created_at"`
	UpdatedAt  Timestamp  `db:"updated_at" json:"updated_at"`
	Permission Permission `db:"permission" json:"-"`
	Email      string     `db:"email" json:"email"`
	Nickname   string     `db:"nickname" json:"nickname"`
	Name       string     `db:"name" json:"name,omitempty"`
	Password   []byte     `db:"password" json:"-"`
}

func (u *User) PasswordMatches(passwd string) bool {
	err := bcrypt.CompareHashAndPassword(u.Password, utils.UnsafeBytes(passwd))
	if err != nil && !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		slog.Error("User: failed to compare hashed password", "error", err)
	}

	return err == nil
}

type UserCreateData struct {
	Email    string `json:"email" validate:"required,email,max=128"`
	Nickname string `json:"nickname" validate:"required,max=32"`
	Name     string `json:"name,omitempty"`
	Password string `json:"password" validate:"required,min=8,max=256"`
}

func NewUser(data UserCreateData, permission Permission, hashCost int) (User, error) {
	now := Timestamp{time.Now().Round(time.Millisecond)}

	if bcrypt.MinCost > hashCost || bcrypt.MaxCost < hashCost {
		hashCost = bcrypt.DefaultCost
	}

	hash, err := bcrypt.GenerateFromPassword(utils.UnsafeBytes(data.Password), hashCost)
	if err != nil {
		return User{}, err
	}

	id := NewSnowflakeTime(now.Time)

	return User{
		ID:         id,
		CreatedAt:  now,
		UpdatedAt:  now,
		Permission: permission,
		Email:      data.Email,
		Nickname:   data.Nickname,
		Name:       data.Name,
		Password:   hash,
	}, nil
}
