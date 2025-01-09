package repository

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/utils/errutils"
)

const (
	_ = 1000 + iota

	CodeUserNotFound
)

var (
	ErrUserNotFound = errutils.NewHttp(
		errors.New("user not found"),
		http.StatusNotFound,
		CodeUserNotFound,
		true,
	)
)

type UserRepository struct {
	q *userQueries
}

func NewUserRepository(db *sqlx.DB) (*UserRepository, error) {
	return &UserRepository{
		q: newUserQueries(db),
	}, nil
}

func (r *UserRepository) Create(ctx context.Context, user dto.User) error {
	sttm, err := r.q.Create()
	if err != nil {
		return err
	}

	_, err = sttm.ExecContext(ctx, user)
	return err
}

func (r *UserRepository) GetById(ctx context.Context, id dto.Snowflake) (dto.User, error) {
	var user dto.User

	sttm, err := r.q.GetById()
	if err != nil {
		return user, err
	}

	if err = sttm.GetContext(ctx, &user, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrUserNotFound
		}
	}
	return user, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (dto.User, error) {
	var user dto.User

	sttm, err := r.q.GetByEmail()
	if err != nil {
		return user, err
	}

	if err = sttm.GetContext(ctx, &user, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrUserNotFound
		}
	}
	return user, err
}

func (r *UserRepository) UpdateName(ctx context.Context, id dto.Snowflake, name string) (dto.User, error) {
	now := time.Now().UnixMilli()

	var user dto.User

	sttm, err := r.q.UpdateName()
	if err != nil {
		return user, err
	}

	name2 := sql.NullString{String: name}
	if name != "" {
		name2.Valid = true
	}

	if err = sttm.GetContext(ctx, &user, name2, now, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrUserNotFound
		}
	}
	return user, err
}

func (r *UserRepository) DeleteById(ctx context.Context, id dto.Snowflake) (dto.User, error) {
	var user dto.User

	sttm, err := r.q.DeleteById()
	if err != nil {
		return user, err
	}

	if err = sttm.GetContext(ctx, &user, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrUserNotFound
		}
	}
	return user, err
}

func (u *UserRepository) Close() error {
	return u.q.Close()
}
