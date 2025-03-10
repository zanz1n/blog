package repository

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/utils/errutils"
)

const (
	_ = 2000 + iota

	CodeAuthTokenExpired
	CodeAuthTokenInvalid
	CodeInvalidRefreshToken
)

var (
	ErrExpiredAuthToken = errutils.NewHttpS(
		"Authentication token expired",
		http.StatusUnauthorized,
		CodeAuthTokenExpired,
		true,
	)

	ErrInvalidAuthToken = errutils.NewHttpS(
		"Authentication token invalid",
		http.StatusUnauthorized,
		CodeAuthTokenInvalid,
		true,
	)

	ErrInvalidRefreshToken = errutils.NewHttpS(
		"Refresh token invalid",
		http.StatusUnauthorized,
		CodeInvalidRefreshToken,
		true,
	)
)

const (
	refreshTokenExpiry = 7 * 24 * time.Hour
	// Number of bytes.
	// Base64 encoded size may vary
	refreshTokenLen = 64
)

var (
	base64d = base64.StdEncoding
	// Size of the refresh token string.
	RefreshTokenLen = base64d.EncodedLen(refreshTokenLen)
)

type AuthRepository struct {
	parser *jwt.Parser

	priv ed25519.PrivateKey
	pub  ed25519.PublicKey

	issuer string

	kv KVStorer
}

func NewAuthRepository(
	priv ed25519.PrivateKey,
	pub ed25519.PublicKey,
	issuer string,
	kv KVStorer,
) *AuthRepository {
	return &AuthRepository{
		parser: jwt.NewParser(),
		priv:   priv,
		pub:    pub,
		issuer: issuer,
		kv:     kv,
	}
}

func (r *AuthRepository) ValidateRefreshToken(
	ctx context.Context,
	token string,
) (dto.Snowflake, error) {
	tokenb, err := decodeRefreshToken(token)
	if err != nil {
		return 0, err
	}

	userId := getRefreshTokenUser(tokenb)
	key := fmt.Sprintf("refresh_token/%s", userId)

	refreshToken, err := r.kv.GetEx(ctx, key, refreshTokenExpiry)
	if err != nil {
		if errors.Is(err, ErrValueNotFound) {
			err = ErrInvalidRefreshToken
		}
		return 0, err
	}

	if refreshToken != token {
		return 0, ErrInvalidRefreshToken
	}

	return userId, nil
}

func (r *AuthRepository) GenRefreshToken(
	ctx context.Context,
	userId dto.Snowflake,
) (string, error) {
	key := fmt.Sprintf("refresh_token/%s", userId)

	refreshToken, err := r.kv.GetEx(ctx, key, refreshTokenExpiry)
	if err == nil {
		return refreshToken, nil
	}

	if !errors.Is(err, ErrValueNotFound) {
		return "", err
	}

	refreshToken = generateRefreshToken(userId)

	err = r.kv.SetEx(ctx, key, refreshToken, refreshTokenExpiry)
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (r *AuthRepository) DeleteRefreshTokens(
	ctx context.Context,
	userId dto.Snowflake,
) error {
	key := fmt.Sprintf("refresh_token/%s", userId)

	err := r.kv.Delete(ctx, key)
	if errors.Is(err, ErrValueNotFound) {
		err = nil
	}

	return err
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

func decodeRefreshToken(ts string) ([]byte, error) {
	token, err := base64d.DecodeString(ts)
	if err != nil || len(token) != refreshTokenLen {
		err = ErrInvalidRefreshToken
	}

	return token, err
}

func generateRefreshToken(userId dto.Snowflake) string {
	token := make([]byte, refreshTokenLen)
	rand.Read(token)

	binary.LittleEndian.PutUint64(token, uint64(userId))

	return base64d.EncodeToString(token)
}

func getRefreshTokenUser(token []byte) dto.Snowflake {
	return dto.Snowflake(binary.LittleEndian.Uint64(token[0:8]))
}
