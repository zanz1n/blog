package repository

import (
	"strings"

	"github.com/jmoiron/sqlx"
)

const kvExistsQuery = `SELECT COUNT(1)
FROM keyvalue
WHERE key = $1 AND (expiry IS NULL OR expiry > $2)`

const kvGetQuery = `SELECT value
FROM keyvalue 
WHERE key = $1 AND (expiry IS NULL OR expiry > $2)`

const kvGetExQuery = `UPDATE keyvalue
SET expiry = $1
WHERE key = $2 AND (expiry IS NULL OR expiry > $3)
RETURNING value`

const kvSetQueryPG = `INSERT INTO keyvalue
(key, value, expiry) VALUES ($1, $2, $3)
ON CONFLICT (key) DO UPDATE
SET value = $2, expiry = $3`

const kvSetQuerySQLITE = `INSERT OR REPLACE INTO keyvalue
(key, value, expiry) VALUES ($1, $2, $3)`

const kvDeleteQuery = `DELETE FROM keyvalue
WHERE key = $1 AND (expiry IS NULL OR expiry > $2)`

const kvCleanupQuery = `DELETE FROM keyvalue expiry < $1`

type kvQueries struct {
	*queries
}

func newKvQueries(db *sqlx.DB) kvQueries {
	q := newQueries(db, "KVQueries")

	q.add(kvExistsQuery, "Exists")

	q.add(kvGetQuery, "Get")
	q.add(kvGetExQuery, "GetEx")

	q.add(kvDeleteQuery, "Delete")
	q.add(kvCleanupQuery, "Cleanup")

	if strings.Contains(db.DriverName(), "sqlite") {
		q.add(kvSetQuerySQLITE, "Set")
	} else {
		q.add(kvSetQueryPG, "Set")
	}

	return kvQueries{q}
}

func (q *kvQueries) Exists() (*sqlx.Stmt, error) {
	return q.get("Exists")
}

func (q *kvQueries) Get() (*sqlx.Stmt, error) {
	return q.get("Get")
}

func (q *kvQueries) GetEx() (*sqlx.Stmt, error) {
	return q.get("GetEx")
}

func (q *kvQueries) Set() (*sqlx.Stmt, error) {
	return q.get("Set")
}

func (q *kvQueries) Delete() (*sqlx.Stmt, error) {
	return q.get("Delete")
}

func (q *kvQueries) Cleanup() (*sqlx.Stmt, error) {
	return q.get("Cleanup")
}
