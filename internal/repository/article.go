package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/utils/errutils"
)

const (
	_ = 2000 + iota

	CodeArticleNotFound
	CodeArticleAlreadyExists
)

var (
	ErrArticleNotFound = errutils.NewHttp(
		errors.New("article not found"),
		http.StatusNotFound,
		CodeArticleNotFound,
		true,
	)
	ErrArticleAlreadyExists = errutils.NewHttp(
		errors.New("article already exists"),
		http.StatusConflict,
		CodeArticleAlreadyExists,
		true,
	)
)

type ArticleRepository struct {
	q  articleQueries
	ur *UserRepository
}

func NewArticleRepository(db *sqlx.DB, userRepo *UserRepository) *ArticleRepository {
	return &ArticleRepository{
		q:  newArticleQueries(db),
		ur: userRepo,
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

func (r *ArticleRepository) GetFull(ctx context.Context, id dto.Snowflake) (dto.Article, error) {
	return r.getAnyWithUser(ctx, id, "GetFull")
}

func (r *ArticleRepository) GetMany(
	ctx context.Context,
	pag dto.Pagination,
) ([]dto.Article, error) {
	var articles []dto.Article

	sttm, err := r.q.GetMany()
	if err != nil {
		return articles, err
	}

	err = sttm.SelectContext(ctx, &articles, pag.Limit, pag.Offset)
	if err != nil {
		slog.Error("ArticleRepository: GetMany: sql error", "error", err)
	}
	return articles, err
}

func (r *ArticleRepository) GetManyByUser(
	ctx context.Context,
	userId dto.Snowflake,
	pag dto.Pagination,
) ([]dto.Article, error) {
	var articles []dto.Article

	user, err := r.ur.GetById(ctx, userId)
	if err != nil {
		return nil, err
	}

	sttm, err := r.q.GetManyByUser()
	if err != nil {
		return nil, err
	}

	err = sttm.SelectContext(ctx, &articles, userId, pag.Limit, pag.Offset)
	if err != nil {
		slog.Error("ArticleRepository: GetManyByUser: sql error", "error", err)
	}

	for i := 0; i < len(articles); i++ {
		articles[i].User = &user
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
) (dto.Article, error) {
	now := time.Now().UnixMilli()

	var article dto.Article

	sttm, err := r.q.UpdateContent()
	if err != nil {
		return article, err
	}

	err = sttm.GetContext(ctx, &article, idx, content, now, id)
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
