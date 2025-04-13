package repository

import (
	"github.com/jmoiron/sqlx"
)

const userCreateQuery = `INSERT INTO users VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

const userGetByIdQuery = `SELECT * FROM users WHERE id = $1`

const userGetByEmailQuery = `SELECT * FROM users WHERE email = $1`

const userUpdateNameQuery = `UPDATE users SET name = $1, updated_at = $2 WHERE id = $3 RETURNING *`

const userDeleteByIdQuery = `DELETE FROM users WHERE id = $1 RETURNING *`

type userQueries struct {
	*Queries
}

func newUserQueries(db *sqlx.DB) userQueries {
	q := NewQueries(db, "UserQueries")

	q.Add(userCreateQuery, "Create")
	q.Add(userGetByIdQuery, "GetById")
	q.Add(userGetByEmailQuery, "GetByEmail")
	q.Add(userUpdateNameQuery, "UpdateName")
	q.Add(userDeleteByIdQuery, "DeleteById")

	return userQueries{q}
}

func (q *userQueries) Create() (*sqlx.Stmt, error) {
	return q.Get("Create")
}

func (q *userQueries) GetById() (*sqlx.Stmt, error) {
	return q.Get("GetById")
}

func (q *userQueries) GetByEmail() (*sqlx.Stmt, error) {
	return q.Get("GetByEmail")
}

func (q *userQueries) UpdateName() (*sqlx.Stmt, error) {
	return q.Get("UpdateName")
}

func (q *userQueries) DeleteById() (*sqlx.Stmt, error) {
	return q.Get("DeleteById")
}
