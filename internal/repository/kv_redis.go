package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/valkey-io/valkey-go"
	"github.com/zanz1n/blog/internal/utils"
)

var _ KVStorer = &RedisKV{}

type RedisKV struct {
	c valkey.Client
}

func NewRedisKV(client valkey.Client) *RedisKV {
	return &RedisKV{c: client}
}

// Exists implements KVStorer.
func (r *RedisKV) Exists(ctx context.Context, key string) (bool, error) {
	cmd := r.c.B().Exists().Key(key).Build()
	return r.c.Do(ctx, cmd).AsBool()
}

// Get implements KVStorer.
func (r *RedisKV) Get(ctx context.Context, key string) (string, error) {
	cmd := r.c.B().Get().Key(key).Build()
	value, err := r.c.Do(ctx, cmd).ToString()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			err = ErrValueNotFound
		}
	}

	return value, err
}

// GetEx implements KVStorer.
func (r *RedisKV) GetEx(
	ctx context.Context,
	key string,
	ttl time.Duration,
) (string, error) {
	cmd := r.c.B().Getex().Key(key).Ex(ttl).Build()
	value, err := r.c.Do(ctx, cmd).ToString()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			err = ErrValueNotFound
		}
	}

	return value, err
}

// GetValue implements KVStorer.
func (r *RedisKV) GetValue(ctx context.Context, key string, v any) error {
	value, err := r.Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal(utils.UnsafeBytes(value), v)
}

// GetValueEx implements KVStorer.
func (r *RedisKV) GetValueEx(
	ctx context.Context,
	key string,
	ttl time.Duration,
	v any,
) error {
	value, err := r.GetEx(ctx, key, ttl)
	if err != nil {
		return err
	}

	return json.Unmarshal(utils.UnsafeBytes(value), v)
}

// Set implements KVStorer.
func (r *RedisKV) Set(ctx context.Context, key string, value string) error {
	cmd := r.c.B().Set().Key(key).Value(value).Build()
	return r.c.Do(ctx, cmd).Error()
}

// SetEx implements KVStorer.
func (r *RedisKV) SetEx(
	ctx context.Context,
	key string,
	value string,
	ttl time.Duration,
) error {
	cmd := r.c.B().Set().Key(key).Value(value).Ex(ttl).Build()
	return r.c.Do(ctx, cmd).Error()
}

// SetValue implements KVStorer.
func (r *RedisKV) SetValue(ctx context.Context, key string, v any) error {
	value, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return r.Set(ctx, key, utils.UnsafeString(value))
}

// SetValueEx implements KVStorer.
func (r *RedisKV) SetValueEx(
	ctx context.Context,
	key string,
	v any,
	ttl time.Duration,
) error {
	value, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return r.SetEx(ctx, key, utils.UnsafeString(value), ttl)
}

// Delete implements KVStorer.
func (r *RedisKV) Delete(ctx context.Context, key string) error {
	cmd := r.c.B().Del().Key(key).Build()
	ct, err := r.c.Do(ctx, cmd).AsInt64()
	if err != nil {
		return err
	}

	if ct == 0 {
		return ErrValueNotFound
	}
	return nil
}
