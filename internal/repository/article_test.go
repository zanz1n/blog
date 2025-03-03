package repository_test

import (
	"context"
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/repository"
)

func articleRepo(t *testing.T) (*repository.ArticleRepository, *repository.UserRepository) {
	db := GetDb(t)
	repo := repository.NewArticleRepository(db)
	userRepo := repository.NewUserRepository(db)
	return repo, userRepo
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
		t.Parallel()
		_, err := articles.Get(context.Background(), dto.NewSnowflake())
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrArticleNotFound)
	})

	article, user := createArticle(t, articles, users)
	user.Password = nil

	t.Run("Duplicate", func(t *testing.T) {
		err := articles.Create(context.Background(), article)
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

func TestArticleUpdate(t *testing.T) {
	t.Parallel()
	articles, users := articleRepo(t)

	t.Run("Inexistent", func(t *testing.T) {
		t.Parallel()
		data := articleData()
		_, err := articles.UpdateData(
			context.Background(),
			dto.NewSnowflake(),
			data.Title,
			data.Description,
		)

		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrArticleNotFound)
	})

	article, _ := createArticle(t, articles, users)

	time.Sleep(5 * time.Millisecond)

	t.Run("UpdateData", func(t *testing.T) {
		newData := articleData()

		time.Sleep(5 * time.Millisecond)

		article2, err := articles.UpdateData(
			context.Background(),
			article.ID,
			newData.Title,
			newData.Description,
		)
		assert.NoError(t, err)

		assert.Equal(t, newData.Title, article2.Title)
		assert.Equal(t, newData.Description, article2.Description)

		assert.Greater(t,
			article2.UpdatedAt.UnixMilli(),
			article.UpdatedAt.UnixMilli(),
		)

		article.UpdatedAt = article2.UpdatedAt
		article.Title = newData.Title
		article.Description = newData.Description

		assert.Nil(t, article2.Indexing)
		article2.Indexing = article.Indexing

		assert.Nil(t, article2.Content)
		article2.Content = article.Content

		assert.Equal(t, article, article2)
	})

	t.Run("Fetch1", func(t *testing.T) {
		article2, err := articles.GetWithContent(context.Background(), article.ID)
		assert.NoError(t, err)
		assert.Equal(t, article, article2)
	})

	t.Run("UpdateContent", func(t *testing.T) {
		articleIdx, articleContent, _ := articleData2()

		time.Sleep(5 * time.Millisecond)

		article2, err := articles.UpdateContent(
			context.Background(),
			article.ID,
			articleIdx,
			articleContent,
		)
		assert.NoError(t, err)

		article2.Indexing = articleIdx
		article2.Content = articleContent

		assert.Greater(t,
			article2.UpdatedAt.UnixMilli(),
			article.UpdatedAt.UnixMilli(),
		)
		article.UpdatedAt = article2.UpdatedAt
		article.Indexing = articleIdx
		article.Content = articleContent

		assert.Equal(t, article, article2)
	})

	t.Run("Fetch2", func(t *testing.T) {
		article2, err := articles.GetWithContent(context.Background(), article.ID)
		assert.NoError(t, err)
		assert.Equal(t, article, article2)
	})
}

func TestArticleDelete(t *testing.T) {
	t.Parallel()
	articles, users := articleRepo(t)

	t.Run("Inexistent", func(t *testing.T) {
		t.Parallel()
		_, err := articles.Delete(context.Background(), dto.NewSnowflake())
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrArticleNotFound)
	})

	article, _ := createArticle(t, articles, users)

	t.Run("Delete", func(t *testing.T) {
		article2, err := articles.Delete(context.Background(), article.ID)
		assert.NoError(t, err)

		assert.Nil(t, article2.Indexing)
		assert.Nil(t, article2.Content)

		article2.Indexing = article.Indexing
		article2.Content = article.Content
	})

	t.Run("Fetch", func(t *testing.T) {
		_, err := articles.Get(context.Background(), article.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrArticleNotFound)
	})
}

func TestArticleGetMany(t *testing.T) {
	const (
		UserCount = 3
		Count     = 13
		PageSize  = 5
	)

	t.Parallel()

	db, err := InitDb(t)
	assert.NoError(t, err)

	userRepo := repository.NewUserRepository(db)
	articleRepo := repository.NewArticleRepository(db)

	users := make([]dto.User, UserCount)
	articles := make([]dto.Article, UserCount*Count)
	articlesByUser := make([][]dto.Article, UserCount)

	for u := 0; u < UserCount; u++ {
		user, err := dto.NewUser(userData(), dto.PermissionDefault, 4)
		assert.NoError(t, err)

		users[u] = user
		articlesByUser[u] = make([]dto.Article, 0, Count)

		for i := 0; i < Count; i++ {
			data := articleData()
			article := dto.NewArticle(user.ID, nil, nil, data)

			articles[(u*Count)+i] = article
			articlesByUser[u] = append(articlesByUser[u], article)
		}
		sortById(articlesByUser[u])
	}
	sortById(articles)

	assert.True(t, t.Run("CreateUsers", func(t *testing.T) {
		for _, user := range users {
			err := userRepo.Create(context.Background(), user)
			assert.NoError(t, err)
		}
	}))

	assert.True(t, t.Run("CreateArticles", func(t *testing.T) {
		for _, article := range articles {
			err := articleRepo.Create(context.Background(), article)
			assert.NoError(t, err)
		}
	}))

	t.Run("GetMany", func(t *testing.T) {
		result := []dto.Article{}

		for i := 0; i < (UserCount*Count/PageSize)+1; i++ {
			articles, err := articleRepo.GetMany(
				context.Background(),
				dto.Pagination{
					Limit:  PageSize,
					Offset: i * PageSize,
				},
			)
			assert.NoError(t, err)

			result = append(result, articles...)
		}

		assert.Equal(t, articles, result)
	})

	t.Run("GetManyByUser", func(t *testing.T) {
		for u := 0; u < UserCount; u++ {
			result := []dto.Article{}

			for i := 0; i < (Count/PageSize)+1; i++ {
				articles, err := articleRepo.GetManyByUser(
					context.Background(),
					users[u].ID,
					dto.Pagination{
						Limit:  PageSize,
						Offset: i * PageSize,
					},
				)
				assert.NoError(t, err)

				result = append(result, articles...)
			}

			assert.Equal(t, articlesByUser[u], result)
		}
	})
}

func sortById(s []dto.Article) {
	slices.SortFunc(s, func(a, b dto.Article) int {
		if b.ID > a.ID {
			return -1
		} else if a.ID > b.ID {
			return 1
		}
		return 0
	})
}

func createArticle(
	t *testing.T,
	articles *repository.ArticleRepository,
	users *repository.UserRepository,
) (dto.Article, dto.User) {
	user, err := dto.NewUser(userData(), dto.PermissionDefault, 4)
	assert.NoError(t, err)

	assert.True(t, t.Run("CreateUser", func(t *testing.T) {
		err := users.Create(context.Background(), user)
		assert.NoError(t, err)
	}))

	articleIdx, articleContent, data := articleData2()
	article := dto.NewArticle(user.ID, articleIdx, articleContent, data)

	assert.Equal(t, article.Content, articleContent)
	assert.Equal(t, articleIdx, article.Indexing)
	assert.Equal(t, data.Title, article.Title)
	assert.Equal(t, data.Description, article.Description)

	assert.True(t, t.Run("Create", func(t *testing.T) {
		err := articles.Create(context.Background(), article)
		assert.NoError(t, err)
	}))

	return article, user
}
