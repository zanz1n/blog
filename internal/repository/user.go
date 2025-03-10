package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/utils/errutils"
)

const (
	_ = 1000 + iota

	CodeUserNotFound
	CodeUserAlreadyExists
)

var (
	ErrUserNotFound = errutils.NewHttpS(
		"User not found",
		http.StatusNotFound,
		CodeUserNotFound,
		true,
	)
	ErrUserAlreadyExists = errutils.NewHttpS(
		"User already exists, try a different email address",
		http.StatusConflict,
		CodeUserAlreadyExists,
		true,
	)
)

type UserRepository struct {
	q userQueries
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{q: newUserQueries(db)}
}

func (r *UserRepository) Create(ctx context.Context, user dto.User) error {
	sttm, err := r.q.Create()
	if err != nil {
		return err
	}

	name2 := sql.NullString{String: user.Name}
	if user.Name != "" {
		name2.Valid = true
	}

	_, err = sttm.ExecContext(ctx,
		user.ID,
		user.CreatedAt,
		user.UpdatedAt,
		user.Permission,
		user.Email,
		user.Nickname,
		name2,
		user.Password,
	)
	if err != nil {
		if isUniqueConstraintViolation(err) {
			err = ErrUserAlreadyExists
		} else {
			slog.Error("UserRepository: Create: sql error", "error", err)
		}
	}
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
		} else {
			slog.Error("UserRepository: GetById: sql error", "error", err)
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
		} else {
			slog.Error("UserRepository: GetByEmail: sql error", "error", err)
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
		} else {
			slog.Error("UserRepository: UpdateName: sql error", "error", err)
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
		} else {
			slog.Error("UserRepository: DeleteById: sql error", "error", err)
		}
	}
	return user, err
}

func (r *UserRepository) Close() error {
	return r.q.Close()
}
