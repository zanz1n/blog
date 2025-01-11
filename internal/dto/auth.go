package dto

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	PermissionReadPosts Permission = 1 << iota
	PermissionWritePosts
	PermissionReadComments
	PermissionWriteComments
	PermissionModerateAllComments
	PermissionReadProfiles
	PermissionWriteProfiles

	PermissionVisitor = PermissionReadPosts |
		PermissionReadComments |
		PermissionReadProfiles
	PermissionDefault   = PermissionVisitor | PermissionWriteComments
	PermisisonPublisher = PermissionDefault | PermissionWritePosts
	PermissionModerator = PermissionDefault | PermissionModerateAllComments
)

var _ fmt.Stringer = PermissionDefault

type Permission int

// String implements fmt.Stringer.
func (p Permission) String() string {
	return strconv.Itoa(int(p))
}

var _ jwt.Claims = &AuthToken{}

type AuthToken struct {
	ID         Snowflake       `json:"sub"`
	IssuedAt   jwt.NumericDate `json:"iat"`
	ExpiresAt  jwt.NumericDate `json:"exp"`
	Issuer     string          `json:"iss"`
	Nickname   string          `json:"nick"`
	Email      string          `json:"email"`
	Permission Permission      `json:"perm"`
}

func NewAuthToken(user *User, issuer string, exp time.Duration) AuthToken {
	now := time.Now().Round(time.Second)

	return AuthToken{
		ID:         user.ID,
		IssuedAt:   jwt.NumericDate{Time: now},
		ExpiresAt:  jwt.NumericDate{Time: now.Add(exp)},
		Issuer:     issuer,
		Nickname:   user.Nickname,
		Email:      user.Email,
		Permission: user.Permission,
	}
}

// GetAudience implements jwt.Claims.
func (t *AuthToken) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

// GetExpirationTime implements jwt.Claims.
func (t *AuthToken) GetExpirationTime() (*jwt.NumericDate, error) {
	return &t.ExpiresAt, nil
}

// GetIssuedAt implements jwt.Claims.
func (t *AuthToken) GetIssuedAt() (*jwt.NumericDate, error) {
	return &t.IssuedAt, nil
}

// GetIssuer implements jwt.Claims.
func (t *AuthToken) GetIssuer() (string, error) {
	return t.Issuer, nil
}

// GetNotBefore implements jwt.Claims.
func (t *AuthToken) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

// GetSubject implements jwt.Claims.
func (t *AuthToken) GetSubject() (string, error) {
	return "", nil
}
