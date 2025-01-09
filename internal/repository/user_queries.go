package repository

import (
	"io"
	"sync"
	"sync/atomic"

	"github.com/jmoiron/sqlx"
)

const userCreateQuery = `INSERT INTO users (
    id,
    created_at,
    updated_at,
    permission,
    email,
    nickname,
    name,
    password
)
VALUES (
	:id,
	:created_at,
	:updated_at,
	:permission,
	:email,
	:nickname,
	:name,
	:password
);`

const userGetByIdQuery = `SELECT * FROM users WHERE id = $1`

const userGetByEmailQuery = `SELECT * FROM users WHERE email = $1`

const userUpdateNameQuery = `UPDATE users SET name = $1, updated_at = $2 WHERE id = $3 RETURNING *`

const userDeleteByIdQuery = `DELETE FROM users WHERE id = $1 RETURNING *`

type userQueries struct {
	db *sqlx.DB

	create     atomic.Pointer[sqlx.NamedStmt]
	getById    atomic.Pointer[sqlx.Stmt]
	getByEmail atomic.Pointer[sqlx.Stmt]
	updateName atomic.Pointer[sqlx.Stmt]
	deleteById atomic.Pointer[sqlx.Stmt]

	closers   []io.Closer
	closersMu sync.Mutex
}

func newUserQueries(db *sqlx.DB) *userQueries {
	return &userQueries{
		db:        db,
		closers:   make([]io.Closer, 0, 5),
		closersMu: sync.Mutex{},
	}
}

func (q *userQueries) Create() (*sqlx.NamedStmt, error) {
	if q.create.Load() == nil {
		create, err := q.db.PrepareNamed(userCreateQuery)
		if err != nil {
			return nil, err
		}
		q.create.Store(create)
		q.append(create)
	}
	return q.create.Load(), nil
}

func (q *userQueries) GetById() (*sqlx.Stmt, error) {
	if q.getById.Load() == nil {
		getById, err := q.db.Preparex(userGetByIdQuery)
		if err != nil {
			return nil, err
		}
		q.getById.Store(getById)
		q.append(getById)
	}
	return q.getById.Load(), nil
}

func (q *userQueries) GetByEmail() (*sqlx.Stmt, error) {
	if q.getByEmail.Load() == nil {
		getByEmail, err := q.db.Preparex(userGetByEmailQuery)
		if err != nil {
			return nil, err
		}
		q.getByEmail.Store(getByEmail)
		q.append(getByEmail)
	}
	return q.getByEmail.Load(), nil
}

func (q *userQueries) UpdateName() (*sqlx.Stmt, error) {
	if q.updateName.Load() == nil {
		updateName, err := q.db.Preparex(userUpdateNameQuery)
		if err != nil {
			return nil, err
		}
		q.updateName.Store(updateName)
		q.append(updateName)
	}
	return q.updateName.Load(), nil
}

func (q *userQueries) DeleteById() (*sqlx.Stmt, error) {
	if q.deleteById.Load() == nil {
		deleteById, err := q.db.Preparex(userDeleteByIdQuery)
		if err != nil {
			return nil, err
		}
		q.deleteById.Store(deleteById)
		q.append(deleteById)
	}
	return q.deleteById.Load(), nil
}

func (q *userQueries) append(c io.Closer) {
	q.closersMu.Lock()
	defer q.closersMu.Unlock()

	q.closers = append(q.closers, c)
}

func (q *userQueries) Close() error {
	q.closersMu.Lock()
	defer q.closersMu.Unlock()

	var lastErr error
	for _, q := range q.closers {
		if err := q.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
