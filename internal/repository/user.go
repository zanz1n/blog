package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/utils/errutils"
)

const (
	_ = 1000 + iota

	CodeUserNotFound
	CodeUserAlreadyExists
)

var (
	ErrUserNotFound = errutils.NewHttp(
		errors.New("user not found"),
		http.StatusNotFound,
		CodeUserNotFound,
		true,
	)
	ErrUserAlreadyExists = errutils.NewHttp(
		errors.New("user already exists, try a different email address"),
		http.StatusConflict,
		CodeUserAlreadyExists,
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
	sttm, err := r.q.Create.Get()
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

	sttm, err := r.q.GetById.Get()
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

	sttm, err := r.q.GetByEmail.Get()
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

	sttm, err := r.q.UpdateName.Get()
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

	sttm, err := r.q.DeleteById.Get()
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

func (u *UserRepository) Close() error {
	return u.q.Close()
}

func isUniqueConstraintViolation(err error) bool {
	if sqliteErr, ok := err.(sqlite3.Error); ok {
		if sqliteErr.Code == sqlite3.ErrConstraint {
			return true
		} else {
			return false
		}
	} else if pgErr, ok := err.(*pgconn.PgError); ok {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return true
		} else {
			return false
		}
	}
	return false
}
