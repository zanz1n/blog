package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/zanz1n/blog/internal/utils"
)

var _ KVStorer = &SqlKV{}

type SqlKV struct {
	q kvQueries
}

func NewSqlKV(db *sqlx.DB) *SqlKV {
	return &SqlKV{
		q: newKvQueries(db),
	}
}

// Exists implements KVStorer.
func (r *SqlKV) Exists(ctx context.Context, key string) (bool, error) {
	sttm, err := r.q.Exists()
	if err != nil {
		return false, err
	}

	now := time.Now().Unix()

	var ct int
	err = sttm.QueryRowContext(ctx, key, now).Scan(&ct)

	return ct == 1, err
}

// Get implements KVStorer.
func (r *SqlKV) Get(ctx context.Context, key string) (string, error) {
	sttm, err := r.q.Get()
	if err != nil {
		return "", err
	}

	now := time.Now().Unix()

	var value string
	err = sttm.QueryRowContext(ctx, key, now).Scan(&value)

	if errors.Is(err, sql.ErrNoRows) {
		err = ErrValueNotFound
	}

	return value, err
}

// GetEx implements KVStorer.
func (r *SqlKV) GetEx(
	ctx context.Context,
	key string,
	ttl time.Duration,
) (string, error) {
	sttm, err := r.q.GetEx()
	if err != nil {
		return "", err
	}

	now := time.Now()
	exp := now.Add(ttl).Unix()

	var value string
	err = sttm.QueryRowContext(ctx, exp, key, now.Unix()).Scan(&value)

	if errors.Is(err, sql.ErrNoRows) {
		err = ErrValueNotFound
	}

	return value, err
}

// GetValue implements KVStorer.
func (r *SqlKV) GetValue(ctx context.Context, key string, v any) error {
	value, err := r.Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal(utils.UnsafeBytes(value), v)
}

// GetValueEx implements KVStorer.
func (r *SqlKV) GetValueEx(
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
func (r *SqlKV) Set(ctx context.Context, key string, value string) error {
	sttm, err := r.q.Set()
	if err != nil {
		return err
	}

	_, err = sttm.ExecContext(ctx, key, value, nil)
	return err
}

// SetEx implements KVStorer.
func (r *SqlKV) SetEx(
	ctx context.Context,
	key string,
	value string,
	ttl time.Duration,
) error {
	sttm, err := r.q.Set()
	if err != nil {
		return err
	}

	exp := time.Now().Add(ttl).Unix()
	_, err = sttm.ExecContext(ctx, key, value, exp)
	return err
}

// SetValue implements KVStorer.
func (r *SqlKV) SetValue(ctx context.Context, key string, v any) error {
	value, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return r.Set(ctx, key, utils.UnsafeString(value))
}

// SetValueEx implements KVStorer.
func (r *SqlKV) SetValueEx(
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
func (r *SqlKV) Delete(ctx context.Context, key string) error {
	sttm, err := r.q.Delete()
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	res, err := sttm.ExecContext(ctx, key, now)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		err = ErrValueNotFound
	}
	return err
}

// Purges expired rows from the KV column.
//
// Expired columns are never returned in get method, but expired
// records are kept danling until cleaned up.
func (r *SqlKV) Cleanup(ctx context.Context) error {
	sttm, err := r.q.Cleanup()
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	_, err = sttm.ExecContext(ctx, now)
	return err
}
