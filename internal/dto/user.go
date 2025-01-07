package dto

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID         Snowflake  `db:"id" json:"id,string"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
	Permission Permission `db:"permission" json:"permission"`
	Email      string     `db:"email" json:"email"`
	Nickname   string     `db:"nickname" json:"nickname"`
	Name       string     `db:"name,omitempty" json:"name"`
	Password   []byte     `db:"password" json:"-"`
}

func (u *User) PasswordMatches(passwd string) bool {
	err := bcrypt.CompareHashAndPassword(u.Password, []byte(passwd))
	return err == nil
}

type UserCreateData struct {
	Email    string `json:"email" validate:"required,email,max=128"`
	Nickname string `json:"nickname" validate:"required,max=32"`
	Name     string `json:"name,omitempty"`
	Password string `json:"password" validate:"required,min=8,max=256"`
}

func NewUser(data UserCreateData, hashCost int) (User, error) {
	now := time.Now()
	id := NewSnowflake()

	if bcrypt.MinCost > hashCost || bcrypt.MaxCost < hashCost {
		hashCost = bcrypt.DefaultCost
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(data.Password), hashCost)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		Email:     data.Email,
		Nickname:  data.Nickname,
		Name:      data.Name,
		Password:  hash,
	}, nil
}
