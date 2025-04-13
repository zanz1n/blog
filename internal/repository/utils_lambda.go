//go:build lambda
// +build lambda

package repository

import (
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func isUniqueConstraintViolation(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return true
		} else {
			return false
		}
	}
	return false
}
