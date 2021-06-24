package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestUsedKeys_Exists(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	k := "existed_test_key"

	mt.Run("exists", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{{"ok", 1}})
		repo, err := NewUsedKeys(mt.DB)
		require.NoError(mt, err)

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"_id", k},
		}))
		ok, err := repo.Exists(context.Background(), k)
		assert.NoError(mt, err)
		assert.True(mt, ok)
	})

	mt.Run("error", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{{"ok", 1}})
		repo, err := NewUsedKeys(mt.DB)
		require.NoError(mt, err)

		mt.AddMockResponses(bson.D{{"ok", 0}})
		ok, err := repo.Exists(context.Background(), k)
		assert.Error(mt, err)
		assert.False(mt, ok)
	})

	mt.Run("not exists", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{{"ok", 1}})
		repo, err := NewUsedKeys(mt.DB)
		require.NoError(mt, err)

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "foo.bar", mtest.FirstBatch))
		ok, err := repo.Exists(context.Background(), k)
		assert.NoError(mt, err)
		assert.False(mt, ok)
	})
}

func TestUsedKeys_Store(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	k := "test_key"

	mt.Run("success", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{{"ok", 1}})
		repo, err := NewUsedKeys(mt.DB)
		require.NoError(mt, err)

		mt.AddMockResponses(mtest.CreateSuccessResponse())
		ok, err := repo.Store(context.Background(), k, time.Second)
		assert.NoError(mt, err)
		assert.True(mt, ok)
	})

	mt.Run("duplicate key", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{{"ok", 1}})
		repo, err := NewUsedKeys(mt.DB)
		require.NoError(mt, err)

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   1,
			Code:    11000,
			Message: "duplicate key error",
		}))
		ok, err := repo.Store(context.Background(), k, time.Second)
		assert.NoError(mt, err)
		assert.False(mt, ok)
	})

	mt.Run("error", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{{"ok", 1}})
		repo, err := NewUsedKeys(mt.DB)
		require.NoError(mt, err)

		mt.AddMockResponses(bson.D{{"ok", 0}})
		ok, err := repo.Store(context.Background(), k, time.Second)
		assert.Error(mt, err)
		assert.False(mt, ok)
	})
}
