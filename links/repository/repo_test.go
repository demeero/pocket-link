package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"github.com/demeero/pocket-link/links/service"
)

func TestRepository_Create(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	l := service.Link{
		Shortened: "shortened_test",
		Original:  "original_test",
		ExpAt:     time.Now().Add(time.Hour).UTC(),
	}

	mt.Run("success", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{{"ok", 1}})
		repo, err := New(mt.DB)
		require.NoError(mt, err)

		mt.AddMockResponses(mtest.CreateSuccessResponse())
		actual, err := repo.Create(context.Background(), l)
		assert.NoError(mt, err)
		assert.Equal(mt, l.Shortened, actual.Shortened)
		assert.Equal(mt, l.Original, actual.Original)
		assert.Equal(mt, l.ExpAt, actual.ExpAt)
		assert.NotZero(mt, actual.CreatedAt)
	})

	mt.Run("error", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{{"ok", 1}})
		repo, err := New(mt.DB)
		require.NoError(mt, err)

		mt.AddMockResponses(bson.D{{"ok", 0}})
		actual, err := repo.Create(context.Background(), l)
		assert.Error(mt, err)
		assert.Zero(mt, actual)
	})
}

func TestRepository_LoadByID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	shortened := "shortened_test"
	expAt := time.Now().Add(time.Hour).UTC()
	createdAt := time.Now().UTC()
	orig := "original_test"

	mt.Run("success", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{{"ok", 1}})
		repo, err := New(mt.DB)
		require.NoError(mt, err)

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"_id", shortened},
			{"original", orig},
			{"exp_at", expAt},
			{"created_at", createdAt},
		}))

		expected := service.Link{
			Shortened: shortened,
			Original:  orig,
			ExpAt:     expAt,
			CreatedAt: createdAt,
		}

		actual, err := repo.LoadByID(context.Background(), shortened)
		assert.NoError(mt, err)
		assert.Equal(mt, expected.Shortened, actual.Shortened)
		assert.Equal(mt, expected.Original, actual.Original)
		assert.Equal(mt, primitive.NewDateTimeFromTime(expected.ExpAt), primitive.NewDateTimeFromTime(actual.ExpAt))
		assert.Equal(mt, primitive.NewDateTimeFromTime(expected.CreatedAt), primitive.NewDateTimeFromTime(actual.CreatedAt))
	})

	mt.Run("error", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{{"ok", 1}})
		repo, err := New(mt.DB)
		require.NoError(mt, err)

		mt.AddMockResponses(bson.D{{"ok", 0}})
		actual, err := repo.LoadByID(context.Background(), shortened)
		assert.Error(mt, err)
		assert.Zero(mt, actual)
	})

	mt.Run("no docs", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{{"ok", 1}})
		repo, err := New(mt.DB)
		require.NoError(mt, err)

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "foo.bar", mtest.FirstBatch))
		actual, err := repo.LoadByID(context.Background(), shortened)
		assert.ErrorIs(mt, err, service.ErrNotFound)
		assert.Zero(mt, actual)
	})
}
