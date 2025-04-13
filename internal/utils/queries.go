package utils

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

var _ io.Closer = &Queries{}

type Queries struct {
	db *sqlx.DB

	name string

	mp map[string]*Lazy[sqlx.Stmt]

	closers   []io.Closer
	closersMu sync.Mutex
}

func NewQueries(db *sqlx.DB, name string) *Queries {
	return &Queries{
		db:   db,
		name: name,
		mp:   make(map[string]*Lazy[sqlx.Stmt]),
	}
}

// This is not thread safe and must be called on initialization.
func (q *Queries) Add(query, name string) {
	lz := NewLazy(func() (*sqlx.Stmt, error) {
		start := time.Now()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		query = strings.TrimSpace(strings.ReplaceAll(query, "\n", " "))

		sttm, err := q.db.PreparexContext(ctx, query)
		if err != nil {
			slog.Error(
				fmt.Sprintf("%s: Failed to prepare query", q.name),
				"name", name,
				TookAttr(start, 10*time.Microsecond),
				"error", err,
			)
			return nil, err
		}

		slog.Info(
			fmt.Sprintf("%s: Prepared query", q.name),
			"name", name,
			TookAttr(start, 10*time.Microsecond),
		)

		q.closersMu.Lock()
		defer q.closersMu.Unlock()
		q.closers = append(q.closers, sttm)
		return sttm, nil
	})

	q.mp[name] = &lz
}

// This is thread safe and can be called at any time.
func (q *Queries) Get(name string) (*sqlx.Stmt, error) {
	lz, ok := q.mp[name]
	if !ok {
		return nil, fmt.Errorf("queries: `%s` query does not exist", name)
	}
	return lz.Get()
}

// Close implements io.Closer.
//
// This is thread safe and can be called at any time.
func (q *Queries) Close() error {
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
