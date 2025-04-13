package kv

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/zanz1n/blog/internal/utils"
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
	*utils.Queries
}

func newKvQueries(db *sqlx.DB) kvQueries {
	q := utils.NewQueries(db, "KVQueries")

	q.Add(kvExistsQuery, "Exists")

	q.Add(kvGetQuery, "Get")
	q.Add(kvGetExQuery, "GetEx")

	q.Add(kvDeleteQuery, "Delete")
	q.Add(kvCleanupQuery, "Cleanup")

	if strings.Contains(db.DriverName(), "sqlite") {
		q.Add(kvSetQuerySQLITE, "Set")
	} else {
		q.Add(kvSetQueryPG, "Set")
	}

	return kvQueries{q}
}

func (q *kvQueries) Exists() (*sqlx.Stmt, error) {
	return q.Get("Exists")
}

func (q *kvQueries) GetQ() (*sqlx.Stmt, error) {
	return q.Get("Get")
}

func (q *kvQueries) GetEx() (*sqlx.Stmt, error) {
	return q.Get("GetEx")
}

func (q *kvQueries) Set() (*sqlx.Stmt, error) {
	return q.Get("Set")
}

func (q *kvQueries) Delete() (*sqlx.Stmt, error) {
	return q.Get("Delete")
}

func (q *kvQueries) Cleanup() (*sqlx.Stmt, error) {
	return q.Get("Cleanup")
}
