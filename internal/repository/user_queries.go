package repository

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/zanz1n/blog/internal/utils"
)

const userCreateQuery = `INSERT INTO users VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

const userGetByIdQuery = `SELECT * FROM users WHERE id = $1`

const userGetByEmailQuery = `SELECT * FROM users WHERE email = $1`

const userUpdateNameQuery = `UPDATE users SET name = $1, updated_at = $2 WHERE id = $3 RETURNING *`

const userDeleteByIdQuery = `DELETE FROM users WHERE id = $1 RETURNING *`

type userQueries struct {
	db *sqlx.DB

	Create     utils.Lazy[sqlx.Stmt]
	GetById    utils.Lazy[sqlx.Stmt]
	GetByEmail utils.Lazy[sqlx.Stmt]
	UpdateName utils.Lazy[sqlx.Stmt]
	DeleteById utils.Lazy[sqlx.Stmt]

	closers   []io.Closer
	closersMu sync.Mutex
}

func newUserQueries(db *sqlx.DB) *userQueries {
	q := &userQueries{
		db:        db,
		closers:   make([]io.Closer, 0, 5),
		closersMu: sync.Mutex{},
	}

	q.Create = utils.NewLazy(q.prepare(userCreateQuery, "Create"))
	q.GetById = utils.NewLazy(q.prepare(userGetByIdQuery, "GetById"))
	q.GetByEmail = utils.NewLazy(q.prepare(userGetByEmailQuery, "GetByEmail"))
	q.UpdateName = utils.NewLazy(q.prepare(userUpdateNameQuery, "UpdateName"))
	q.DeleteById = utils.NewLazy(q.prepare(userDeleteByIdQuery, "DeleteById"))

	return q
}

func (q *userQueries) prepare(query string, name string) func() (*sqlx.Stmt, error) {
	return func() (*sqlx.Stmt, error) {
		start := time.Now()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		sttm, err := q.db.PreparexContext(ctx, query)
		if err != nil {
			slog.Error(
				"UserQueries: Failed to prepare query",
				"name", name,
				utils.TookAttr(start, 10*time.Microsecond),
				"error", err,
			)
			return nil, err
		}

		slog.Info(
			"UserQueries: Prepared query",
			"name", name,
			utils.TookAttr(start, 10*time.Microsecond),
		)

		q.appendCloser(sttm)
		return sttm, nil
	}
}

func (q *userQueries) appendCloser(c io.Closer) {
	q.closersMu.Lock()
	defer q.closersMu.Unlock()
	q.closers = append(q.closers, c)
}

func (q *userQueries) Close() error {
	q.closersMu.Lock()
	defer q.closersMu.Unlock()

	var lastErr error
	for _, c := range q.closers {
		if err := c.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
