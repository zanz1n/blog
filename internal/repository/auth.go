package repository

import (
	"crypto/ed25519"
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/utils/errutils"
)

const (
	_ = 2000 + iota

	CodeAuthTokenExpired
	CodeAuthTokenInvalid
)

var (
	ErrExpiredAuthToken = errutils.NewHttp(
		errors.New("authentication token expired"),
		http.StatusUnauthorized,
		CodeAuthTokenExpired,
		true,
	)

	ErrInvalidAuthToken = errutils.NewHttp(
		errors.New("authentication token invalid"),
		http.StatusUnauthorized,
		CodeAuthTokenInvalid,
		true,
	)
)

type AuthRepository struct {
	parser *jwt.Parser

	priv ed25519.PrivateKey
	pub  ed25519.PublicKey

	issuer string
}

func NewAuthRepository(
	priv ed25519.PrivateKey,
	pub ed25519.PublicKey,
	issuer string,
) *AuthRepository {
	parser := jwt.NewParser(jwt.WithExpirationRequired())

	return &AuthRepository{
		parser: parser,
		priv:   priv,
		pub:    pub,
		issuer: issuer,
	}
}

func (r *AuthRepository) DecodeToken(token string) (dto.AuthToken, error) {
	var claims dto.AuthToken

	t, err := r.parser.ParseWithClaims(token, &claims, r.keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return claims, ErrExpiredAuthToken
		} else {
			return claims, ErrInvalidAuthToken
		}
	}

	if !t.Valid {
		return claims, ErrInvalidAuthToken
	}

	return claims, nil
}

func (r *AuthRepository) EncodeToken(data dto.AuthToken) (string, error) {
	if data.Issuer == "" {
		data.Issuer = r.issuer
	}

	t := jwt.NewWithClaims(jwt.SigningMethodEdDSA, &data)
	return t.SignedString(r.priv)
}

func (r *AuthRepository) keyFunc(t *jwt.Token) (any, error) {
	if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
		return nil, jwt.ErrEd25519Verification
	}

	return r.pub, nil
}
