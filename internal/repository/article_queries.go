package repository

import (
	"github.com/jmoiron/sqlx"
)

const articleCreateQuery = `INSERT INTO articles
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

const articleGetQuery = `SELECT
id, created_at, updated_at, user_id, title, description
FROM articles WHERE id = $1`

const articleGetWithContentQuery = `SELECT
id, created_at, updated_at, user_id, title, description, indexing, content
FROM articles WHERE id = $1`

const articleGetWithRawContentQuery = `SELECT
id, created_at, updated_at, user_id, title, description, raw_content
FROM articles WHERE id = $1`

const articleGetWithUserQuery = `SELECT
articles.id "articles.id",
articles.created_at "articles.created_at",
articles.updated_at "articles.updated_at",
articles.user_id "articles.user_id",
articles.title "articles.title",
articles.description "articles.description",
users.id "users.id",
users.created_at "users.created_at",
users.updated_at "users.updated_at",
users.permission "users.permission",
users.email "users.email",
users.nickname "users.nickname",
users.name "users.name"
FROM articles
INNER JOIN users ON articles.user_id = users.id
WHERE articles.id = $1`

const articleGetFullQuery = `
SELECT
articles.id "articles.id",
articles.created_at "articles.created_at",
articles.updated_at "articles.updated_at",
articles.user_id "articles.user_id",
articles.title "articles.title",
articles.description "articles.description",
articles.indexing "articles.indexing",
articles.content "articles.content",
users.id "users.id",
users.created_at "users.created_at",
users.updated_at "users.updated_at",
users.permission "users.permission",
users.email "users.email",
users.nickname "users.nickname",
users.name "users.name"
FROM articles
INNER JOIN users ON articles.user_id = users.id
WHERE articles.id = $1`

const articleGetMany = `SELECT
articles.id "articles.id",
articles.created_at "articles.created_at",
articles.updated_at "articles.updated_at",
articles.user_id "articles.user_id",
articles.title "articles.title",
articles.description "articles.description",
users.id "users.id",
users.created_at "users.created_at",
users.updated_at "users.updated_at",
users.permission "users.permission",
users.email "users.email",
users.nickname "users.nickname",
users.name "users.name"
FROM articles
INNER JOIN users ON articles.user_id = users.id
WHERE articles.id < $1
ORDER BY articles.id DESC LIMIT $2`

const articleGetManyByUser = `SELECT
id, created_at, updated_at, user_id, title, description
FROM articles
WHERE user_id = $1 AND id < $2
ORDER BY id DESC LIMIT $3`

const articleUpdateDataQuery = `UPDATE articles
SET title = $1, description = $2, updated_at = $3
WHERE id = $4
RETURNING id, created_at, updated_at, user_id, title, description`

const articleUpdateContentQuery = `UPDATE articles
SET indexing = $1, content = $2, raw_content = $3, updated_at = $4
WHERE id = $5
RETURNING id, created_at, updated_at, user_id, title, description`

const articleDeleteQuery = `DELETE FROM articles
WHERE id = $1
RETURNING id, created_at, updated_at, user_id, title, description`

type articleQueries struct {
	*Queries
}

func newArticleQueries(db *sqlx.DB) articleQueries {
	q := NewQueries(db, "ArticleQueries")

	q.Add(articleCreateQuery, "Create")

	q.Add(articleGetQuery, "Get")
	q.Add(articleGetWithUserQuery, "GetWithUser")
	q.Add(articleGetWithContentQuery, "GetWithContent")
	q.Add(articleGetWithRawContentQuery, "GetWithRawContent")
	q.Add(articleGetFullQuery, "GetFull")

	q.Add(articleGetMany, "GetMany")
	q.Add(articleGetManyByUser, "GetManyByUser")

	q.Add(articleUpdateDataQuery, "UpdateData")
	q.Add(articleUpdateContentQuery, "UpdateContent")

	q.Add(articleDeleteQuery, "Delete")

	return articleQueries{q}
}

func (q *articleQueries) Create() (*sqlx.Stmt, error) {
	return q.Get("Create")
}

func (q *articleQueries) GetMany() (*sqlx.Stmt, error) {
	return q.Get("GetMany")
}

func (q *articleQueries) GetManyByUser() (*sqlx.Stmt, error) {
	return q.Get("GetManyByUser")
}

func (q *articleQueries) UpdateData() (*sqlx.Stmt, error) {
	return q.Get("UpdateData")
}

func (q *articleQueries) UpdateContent() (*sqlx.Stmt, error) {
	return q.Get("UpdateContent")
}

func (q *articleQueries) Delete() (*sqlx.Stmt, error) {
	return q.Get("Delete")
}
