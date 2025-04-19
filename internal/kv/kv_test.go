package kv_test

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	assert "github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	valkeyct "github.com/testcontainers/testcontainers-go/modules/valkey"
	"github.com/valkey-io/valkey-go"
	"github.com/zanz1n/blog/internal/kv"
	"github.com/zanz1n/blog/internal/utils"
)

func kvRepo(t *testing.T) kv.KVStorer {
	if testing.Short() {
		db, err := sqlx.Open("sqlite3", "file::memory:")
		assert.NoError(t, err)

		err = utils.MigrateUp(db)
		assert.NoError(t, err)

		db.SetMaxOpenConns(1)
		t.Cleanup(func() {
			db.Close()
		})

		return kv.NewSqlKV(db)
	}

	valkeyCt, err := valkeyct.Run(context.Background(), "valkey/valkey:8-alpine")
	assert.NoError(t, err)

	testcontainers.CleanupContainer(t, valkeyCt)

	valkeyCStr, err := valkeyCt.ConnectionString(context.Background())
	assert.NoError(t, err)

	valkeyUrl, err := valkey.ParseURL(valkeyCStr)
	assert.NoError(t, err)

	client, err := valkey.NewClient(valkeyUrl)
	assert.NoError(t, err)

	return kv.NewRedisKV(client)
}

func TestKv(t *testing.T) {
	t.Parallel()
	repo := kvRepo(t)

	t.Run("SetGet", func(t *testing.T) {
		t.Parallel()

		key := randString(48)
		value := randString(64)

		exists, err := repo.Exists(context.Background(), key)
		assert.NoError(t, err)
		assert.False(t, exists)

		err = repo.Set(context.Background(), key, value)
		assert.NoError(t, err)

		exists, err = repo.Exists(context.Background(), key)
		assert.NoError(t, err)
		assert.True(t, exists)

		value2, err := repo.Get(context.Background(), key)
		assert.NoError(t, err)
		assert.Equal(t, value, value2)
	})

	t.Run("SetGetValue", func(t *testing.T) {
		t.Parallel()

		type valueType struct {
			Field1 string
			Field2 string
			Field3 bool
		}

		key := randString(48)
		value := valueType{
			Field1: randString(32),
			Field2: randString(32),
			Field3: true,
		}

		err := repo.SetValue(context.Background(), key, value)
		assert.NoError(t, err)

		var value2 valueType
		err = repo.GetValue(context.Background(), key, &value2)
		assert.NoError(t, err)
		assert.Equal(t, value2, value)
	})

	t.Run("SetExGet", func(t *testing.T) {
		t.Parallel()

		key := randString(48)
		value := randString(64)

		err := repo.SetEx(context.Background(), key, value, time.Second)
		assert.NoError(t, err)

		value2, err := repo.Get(context.Background(), key)
		assert.NoError(t, err)
		assert.Equal(t, value, value2)

		time.Sleep(2 * time.Second)

		exists, err := repo.Exists(context.Background(), key)
		assert.NoError(t, err)
		assert.False(t, exists)

		_, err = repo.Get(context.Background(), key)
		assert.Error(t, err)
		assert.ErrorIs(t, err, kv.ErrValueNotFound)
	})

	t.Run("SetGetEx", func(t *testing.T) {
		t.Parallel()

		key := randString(48)
		value := randString(64)

		err := repo.Set(context.Background(), key, value)
		assert.NoError(t, err)

		value2, err := repo.GetEx(context.Background(), key, time.Second)
		assert.NoError(t, err)
		assert.Equal(t, value, value2)

		time.Sleep(2 * time.Second)

		_, err = repo.Get(context.Background(), key)
		assert.Error(t, err)
		assert.ErrorIs(t, err, kv.ErrValueNotFound)
	})

	t.Run("Delete", func(t *testing.T) {
		t.Parallel()

		key := randString(48)
		value := randString(64)

		err := repo.Delete(context.Background(), key)
		assert.Error(t, err)
		assert.ErrorIs(t, err, kv.ErrValueNotFound)

		err = repo.Set(context.Background(), key, value)
		assert.NoError(t, err)

		err = repo.Delete(context.Background(), key)
		assert.NoError(t, err)

		_, err = repo.Get(context.Background(), key)
		assert.Error(t, err)
		assert.ErrorIs(t, err, kv.ErrValueNotFound)
	})
}

func randString(n int) string {
	return utils.RandString(n, utils.Alphabet)
}
