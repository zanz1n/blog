package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/utils"
	"github.com/zanz1n/blog/internal/utils/errutils"
)

const (
	_ = 2000 + iota

	CodeArticleNotFound
	CodeArticleAlreadyExists
)

var (
	ErrArticleNotFound = errutils.NewHttpS(
		"Article not found",
		http.StatusNotFound,
		CodeArticleNotFound,
		true,
	)
	ErrArticleAlreadyExists = errutils.NewHttpS(
		"Article already exists",
		http.StatusConflict,
		CodeArticleAlreadyExists,
		true,
	)
)

type ArticleRepository struct {
	q articleQueries
}

func NewArticleRepository(db *sqlx.DB) *ArticleRepository {
	return &ArticleRepository{
		q: newArticleQueries(db),
	}
}

func (r *ArticleRepository) Create(ctx context.Context, article dto.Article) error {
	sttm, err := r.q.Create()
	if err != nil {
		return err
	}

	description2 := sql.NullString{String: article.Description}
	if article.Description != "" {
		description2.Valid = true
	}

	_, err = sttm.ExecContext(ctx,
		article.ID,
		article.CreatedAt,
		article.UpdatedAt,
		article.UserID,
		article.Title,
		description2,
		article.Indexing,
		article.Content,
		utils.UnsafeString(article.RawContent),
	)
	if err != nil {
		if isUniqueConstraintViolation(err) {
			err = ErrArticleAlreadyExists
		} else {
			slog.Error("ArticleRepository: Create: sql error", "error", err)
		}
	}
	return err
}

func (r *ArticleRepository) Get(ctx context.Context, id dto.Snowflake) (dto.Article, error) {
	return r.getAny(ctx, id, "Get")
}

func (r *ArticleRepository) GetWithUser(
	ctx context.Context,
	id dto.Snowflake,
) (dto.Article, error) {
	return r.getAnyWithUser(ctx, id, "GetWithUser")
}

func (r *ArticleRepository) GetWithContent(
	ctx context.Context,
	id dto.Snowflake,
) (dto.Article, error) {
	return r.getAny(ctx, id, "GetWithContent")
}

func (r *ArticleRepository) GetWithRawContent(
	ctx context.Context,
	id dto.Snowflake,
) (dto.Article, error) {
	return r.getAny(ctx, id, "GetWithRawContent")
}

func (r *ArticleRepository) GetFull(ctx context.Context, id dto.Snowflake) (dto.Article, error) {
	return r.getAnyWithUser(ctx, id, "GetFull")
}

func (r *ArticleRepository) GetMany(
	ctx context.Context,
	pag dto.Pagination,
) ([]dto.Article, error) {
	if pag.LastSeen == 0 {
		// math.MaxUint64 results int integer overflow
		pag.LastSeen = math.MaxInt64
	}

	sttm, err := r.q.GetMany()
	if err != nil {
		return nil, err
	}

	rows, err := sttm.QueryxContext(ctx, pag.LastSeen, pag.Limit)
	if err != nil {
		slog.Error("ArticleRepository: GetMany: sql error", "error", err)
	}
	defer rows.Close()

	articles := []dto.Article{}

	for rows.Next() {
		var res struct {
			Article dto.Article `db:"articles"`
			User    dto.User    `db:"users"`
		}

		if err = rows.StructScan(&res); err != nil {
			return nil, err
		}

		res.Article.User = &res.User
		articles = append(articles, res.Article)
	}

	return articles, err
}

func (r *ArticleRepository) GetManyByUser(
	ctx context.Context,
	userId dto.Snowflake,
	pag dto.Pagination,
) ([]dto.Article, error) {
	if pag.LastSeen == 0 {
		// math.MaxUint64 results int integer overflow
		pag.LastSeen = math.MaxInt64
	}

	sttm, err := r.q.GetManyByUser()
	if err != nil {
		return nil, err
	}

	var articles []dto.Article

	err = sttm.SelectContext(ctx, &articles, userId, pag.LastSeen, pag.Limit)
	if err != nil {
		slog.Error("ArticleRepository: GetManyByUser: sql error", "error", err)
	}

	return articles, err
}

func (r *ArticleRepository) UpdateData(
	ctx context.Context,
	id dto.Snowflake,
	title, description string,
) (dto.Article, error) {
	now := time.Now().UnixMilli()

	var article dto.Article

	sttm, err := r.q.UpdateData()
	if err != nil {
		return article, err
	}

	err = sttm.GetContext(ctx, &article, title, description, now, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrArticleNotFound
		} else {
			slog.Error("ArticleRepository: UpdateData: sql error", "error", err)
		}
	}
	return article, err
}

func (r *ArticleRepository) UpdateContent(
	ctx context.Context,
	id dto.Snowflake,
	idx dto.ArticleIndexing,
	content dto.ArticleContent,
	rawContent dto.ArticleRawContent,
) (dto.Article, error) {
	now := time.Now().UnixMilli()

	var article dto.Article

	sttm, err := r.q.UpdateContent()
	if err != nil {
		return article, err
	}

	err = sttm.GetContext(ctx, &article, idx, content, rawContent, now, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrArticleNotFound
		} else {
			slog.Error("ArticleRepository: UpdateContent: sql error", "error", err)
		}
	}
	return article, err
}

func (r *ArticleRepository) Delete(ctx context.Context, id dto.Snowflake) (dto.Article, error) {
	var article dto.Article

	sttm, err := r.q.Delete()
	if err != nil {
		return article, err
	}

	err = sttm.GetContext(ctx, &article, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrArticleNotFound
		} else {
			slog.Error("ArticleRepository: Delete: sql error", "error", err)
		}
	}
	return article, err
}

func (r *ArticleRepository) getAnyWithUser(
	ctx context.Context,
	id dto.Snowflake,
	name string,
) (dto.Article, error) {
	var res struct {
		Article dto.Article `db:"articles"`
		User    dto.User    `db:"users"`
	}

	sttm, err := r.q.get(name)
	if err != nil {
		return dto.Article{}, err
	}

	if err = sttm.GetContext(ctx, &res, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrArticleNotFound
		} else {
			slog.Error(
				fmt.Sprintf("ArticleRepository: %s: sql error", name),
				"error", err,
			)
		}
	}

	article := res.Article
	article.User = &res.User

	return article, err
}

func (r *ArticleRepository) getAny(
	ctx context.Context,
	id dto.Snowflake,
	name string,
) (dto.Article, error) {
	var article dto.Article

	sttm, err := r.q.get(name)
	if err != nil {
		return article, err
	}

	if err = sttm.GetContext(ctx, &article, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrArticleNotFound
		} else {
			slog.Error(
				fmt.Sprintf("ArticleRepository: %s: sql error", name),
				"error", err,
			)
		}
	}
	return article, err
}

func (r *ArticleRepository) Close() error {
	return r.q.Close()
}
