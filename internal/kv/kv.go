package kv

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/zanz1n/blog/internal/utils/errutils"
)

const (
	_ = 5000 + iota

	CodeValueNotFound
)

var (
	ErrValueNotFound = errutils.NewHttpS(
		"Value not found",
		http.StatusInternalServerError,
		CodeValueNotFound,
		false,
	)
)

type KVStorer interface {
	Exists(ctx context.Context, key string) (bool, error)

	Get(ctx context.Context, key string) (string, error)
	GetEx(ctx context.Context, key string, ttl time.Duration) (string, error)
	GetValue(ctx context.Context, key string, v any) error
	GetValueEx(ctx context.Context, key string, ttl time.Duration, v any) error

	Set(ctx context.Context, key string, value string) error
	SetEx(ctx context.Context, key string, value string, ttl time.Duration) error
	SetValue(ctx context.Context, key string, v any) error
	SetValueEx(ctx context.Context, key string, v any, ttl time.Duration) error

	Delete(ctx context.Context, key string) error

	io.Closer
}
