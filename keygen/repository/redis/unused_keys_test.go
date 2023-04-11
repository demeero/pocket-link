package redis

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/demeero/pocket-link/keygen/key"
)

func TestUnusedKeys_LoadAndDelete(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	client.SAdd(context.Background(), unusedSetName, "k1", "k2", "k3")

	uk := NewUnusedKeys(client)
	actual, err := uk.LoadAndDelete(context.Background())
	assert.NoError(t, err)
	assert.NotEmpty(t, actual)

	assert.Equal(t, int64(2), client.SCard(context.Background(), unusedSetName).Val())
}

func TestUnusedKeys_LoadAndDelete_EmptySet(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	uk := NewUnusedKeys(client)
	actual, err := uk.LoadAndDelete(context.Background())
	assert.Equal(t, key.ErrKeyNotFound, err)
	assert.Empty(t, actual)
}

func TestUnusedKeys_LoadAndDelete_RedisError(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mock.ExpectSPop(unusedSetName).SetErr(redis.ErrClosed)

	uk := NewUnusedKeys(db)
	actual, err := uk.LoadAndDelete(context.Background())
	assert.Empty(t, actual)
	assert.Error(t, err)
}

func TestUnusedKeys_Size(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	client.SAdd(context.Background(), unusedSetName, "k1", "k2", "k3")

	uk := NewUnusedKeys(client)
	actual, err := uk.Size(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(3), actual)
}

func TestUnusedKeys_Store(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	k := "test_key"
	uk := NewUnusedKeys(client)
	actual, err := uk.Store(context.Background(), k)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), actual)

	assert.Equal(t, int64(1), client.SCard(context.Background(), unusedSetName).Val())
}

func TestUnusedKeys_StoreDuplicate(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	client.SAdd(context.Background(), unusedSetName, "k1", "k2", "k3")

	k := "k2"
	uk := NewUnusedKeys(client)
	actual, err := uk.Store(context.Background(), k)
	assert.NoError(t, err)
	assert.Zero(t, actual)

	assert.Equal(t, int64(3), client.SCard(context.Background(), unusedSetName).Val())
}

func TestUnusedKeys_Store_RedisError(t *testing.T) {
	db, mock := redismock.NewClientMock()
	k := "k2"
	mock.ExpectSAdd(unusedSetName, k).SetErr(redis.ErrClosed)

	uk := NewUnusedKeys(db)
	actual, err := uk.Store(context.Background())
	assert.Zero(t, actual)
	assert.Error(t, err)
}
