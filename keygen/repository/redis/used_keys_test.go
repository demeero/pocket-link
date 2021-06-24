package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsedKeys_Exists(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	k := "k1"
	client.Set(context.Background(), k, nil, 0)
	uk := NewUsedKeys(client)

	actual, err := uk.Exists(context.Background(), k)
	assert.True(t, actual)
	assert.NoError(t, err)
}

func TestUsedKeys_Exists_NotExists(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	k := "k1"
	client.Set(context.Background(), "k2", nil, 0)
	uk := NewUsedKeys(client)

	actual, err := uk.Exists(context.Background(), k)
	assert.False(t, actual)
	assert.NoError(t, err)
}

func TestUsedKeys_Exists_RedisErr(t *testing.T) {
	db, mock := redismock.NewClientMock()
	k := "k1"
	mock.ExpectExists(k).SetErr(redis.ErrClosed)
	uk := NewUsedKeys(db)

	actual, err := uk.Exists(context.Background(), k)
	assert.False(t, actual)
	assert.Error(t, err)
}

func TestUsedKeys_Store(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	k := "k1"
	uk := NewUsedKeys(client)

	actual, err := uk.Store(context.Background(), k, time.Hour)
	assert.True(t, actual)
	assert.NoError(t, err)

	assert.Equal(t, int64(1), client.Exists(context.Background(), k).Val())
}

func TestUsedKeys_StoreDuplicate(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	k := "k1"
	client.Set(context.Background(), k, nil, 0)

	uk := NewUsedKeys(client)

	actual, err := uk.Store(context.Background(), k, time.Hour)
	assert.False(t, actual)
	assert.NoError(t, err)
}

func TestUsedKeys_Store_RedisErr(t *testing.T) {
	db, mock := redismock.NewClientMock()
	k := "k1"
	mock.ExpectSetNX(k, "", time.Hour).SetErr(redis.ErrClosed)
	uk := NewUsedKeys(db)

	actual, err := uk.Store(context.Background(), k, time.Hour)
	assert.False(t, actual)
	assert.Error(t, err)
}
