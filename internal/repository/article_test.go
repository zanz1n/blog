package repository_test

import (
	"context"
	"math/rand/v2"
	"testing"

	assert "github.com/stretchr/testify/require"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/repository"
)

func articleRepo(t *testing.T) (*repository.ArticleRepository, *repository.UserRepository) {
	db := GetDb(t)
	userRepo := repository.NewUserRepository(db)
	return repository.NewArticleRepository(db, userRepo), userRepo
}

func articleIndexing(n int) dto.ArticleIndexing {
	m := make(dto.ArticleIndexing, n)

	for i := range n {
		m[i] = dto.ArticleIndexingUnit{
			Head: dto.HeadingType(rand.IntN(int(dto.HeadingTypeH4) + 1)),
			Name: randString(12),
			ID:   randString(8),
		}
	}

	return m
}

func articleData2() (dto.ArticleIndexing, dto.ArticleContent, dto.ArticleCreateData) {
	return articleIndexing(4), dto.ArticleContent(randString(256)), articleData()
}

func articleData() dto.ArticleCreateData {
	return dto.ArticleCreateData{
		Title:       randString(64),
		Description: randString(128),
	}
}

func TestArticleCreate(t *testing.T) {
	t.Parallel()
	articles, users := articleRepo(t)

	t.Run("Inexistent", func(t *testing.T) {
		_, err := articles.Get(context.Background(), dto.NewSnowflake())
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrArticleNotFound)
	})

	user, err := dto.NewUser(userData(), dto.PermissionDefault, 4)
	assert.NoError(t, err)

	t.Run("CreateUser", func(t *testing.T) {
		err = users.Create(context.Background(), user)
		assert.NoError(t, err)
	})

	user.Password = nil

	articleIdx, articleContent, data := articleData2()
	article := dto.NewArticle(user.ID, articleIdx, articleContent, data)
	assert.Equal(t, article.Content, articleContent)
	assert.Equal(t, articleIdx, article.Indexing)
	assert.Equal(t, data.Title, article.Title)
	assert.Equal(t, data.Description, article.Description)

	t.Run("Create", func(t *testing.T) {
		err = articles.Create(context.Background(), article)
		assert.NoError(t, err)
	})

	t.Run("Duplicate", func(t *testing.T) {
		err = articles.Create(context.Background(), article)
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrArticleAlreadyExists)
	})

	t.Run("Get", func(t *testing.T) {
		article2, err := articles.Get(context.Background(), article.ID)
		assert.NoError(t, err)

		assert.Nil(t, article2.User)
		assert.Nil(t, article2.Indexing)
		assert.Nil(t, article2.Content)

		article2.Indexing = article.Indexing
		article2.Content = article.Content

		assert.Equal(t, article, article2)
	})

	t.Run("GetWithUser", func(t *testing.T) {
		article2, err := articles.GetWithUser(context.Background(), article.ID)
		assert.NoError(t, err)

		assert.Equal(t, &user, article2.User)
		assert.Nil(t, article2.Indexing)
		assert.Nil(t, article2.Content)

		article2.User = nil
		article2.Indexing = article.Indexing
		article2.Content = article.Content

		assert.Equal(t, article, article2)
	})

	t.Run("GetWithContent", func(t *testing.T) {
		article2, err := articles.GetWithContent(context.Background(), article.ID)
		assert.NoError(t, err)

		assert.Nil(t, article2.User)
		assert.Equal(t, article, article2)
	})

	t.Run("GetFull", func(t *testing.T) {
		article2, err := articles.GetFull(context.Background(), article.ID)
		assert.NoError(t, err)

		assert.Equal(t, &user, article2.User)
		article2.User = nil

		assert.Equal(t, article, article2)
	})
}
