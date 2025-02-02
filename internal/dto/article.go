package dto

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"

	"github.com/a-h/templ"
	"github.com/zanz1n/blog/internal/utils"
)

type HeadingType uint8

const (
	HeadingTypeNone HeadingType = iota
	HeadingTypeH1
	HeadingTypeH2
	HeadingTypeH3
	HeadingTypeH4
)

type Article struct {
	ID        Snowflake `db:"id" json:"id,string"`
	CreatedAt Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt Timestamp `db:"updated_at" json:"updated_at"`
	UserID    Snowflake `db:"user_id" json:"user_id,string"`
	Title     string    `db:"title" json:"title"`

	// Can be nil if not fetched with user
	User *User `db:"user" json:"user,omitempty"`

	// Can be empty if not fetched with content
	Indexing ArticleIndexing `db:"indexing" json:"indexing,omitempty"`
	// Can be empty if not fetched with content
	Content ArticleContent `db:"content" json:"content,omitempty"`
}

var (
	_nullArticleIndexing = ArticleIndexing(nil)

	_ sql.Scanner   = &_nullArticleIndexing
	_ driver.Valuer = _nullArticleIndexing
)

type ArticleIndexingUnit struct {
	Head HeadingType `json:"head"`
	Name string      `json:"name"`
	ID   string      `json:"id"`
}

type ArticleIndexing []ArticleIndexingUnit

// Scan implements sql.Scanner.
func (a *ArticleIndexing) Scan(src any) (err error) {
	switch src := src.(type) {
	case []byte:
		if err = json.Unmarshal(src, a); err != nil {
			err = articleIndexingScanErr(src, err)
		}
	case string:
		err = json.Unmarshal(utils.UnsafeBytes(src), a)
		if err != nil {
			err = articleIndexingScanErr(src, err)
		}
	default:
		err = articleIndexingScanErr(src, nil)
	}
	return
}

// Value implements driver.Valuer.
func (a ArticleIndexing) Value() (driver.Value, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("failed to encode ArticleIndexing: %s", err)
	}

	return utils.UnsafeString(b), nil
}

var (
	_nullArticleContent = ArticleContent(nil)

	_ templ.Component = _nullArticleContent
	_ sql.Scanner     = &_nullArticleContent
	_ driver.Valuer   = _nullArticleContent
)

type ArticleContent []byte

// Render implements templ.Component.
func (c ArticleContent) Render(ctx context.Context, w io.Writer) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, err := w.Write(c)
	return err
}

// Scan implements sql.Scanner.
func (c *ArticleContent) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		*c = src
	case string:
		*c = utils.UnsafeBytes(src)
	case nil:
		*c = nil
	default:
		return fmt.Errorf("Scan: unable to scan type %T into ArticleContent", src)
	}

	return nil
}

// Value implements driver.Valuer.
func (c ArticleContent) Value() (driver.Value, error) {
	return utils.UnsafeString(c), nil
}

func articleIndexingScanErr(src any, err error) error {
	if err == nil {
		return fmt.Errorf("Scan: unable to scan type %T into ArticleIndexing", src)
	}
	return fmt.Errorf("Scan: unable to scan type %T into ArticleIndexing: %s", src, err)
}
