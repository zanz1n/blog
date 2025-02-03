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
	*queries
}

func newUserQueries(db *sqlx.DB) userQueries {
	q := newQueries(db, "UserQueries")

	q.add(userCreateQuery, "Create")
	q.add(userGetByIdQuery, "GetById")
	q.add(userGetByEmailQuery, "GetByEmail")
	q.add(userUpdateNameQuery, "UpdateName")
	q.add(userDeleteByIdQuery, "DeleteById")

	return userQueries{q}
}

func (q *userQueries) Create() (*sqlx.Stmt, error) {
	return q.get("Create")
}

func (q *userQueries) GetById() (*sqlx.Stmt, error) {
	return q.get("GetById")
}

func (q *userQueries) GetByEmail() (*sqlx.Stmt, error) {
	return q.get("GetByEmail")
}

func (q *userQueries) UpdateName() (*sqlx.Stmt, error) {
	return q.get("UpdateName")
}

func (q *userQueries) DeleteById() (*sqlx.Stmt, error) {
	return q.get("DeleteById")
}
