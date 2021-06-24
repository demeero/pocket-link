package key

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestKeys_Use(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usedRepo := NewMockUsedKeysRepository(ctrl)
	unusedRepo := NewMockUnusedKeysRepository(ctrl)
	testKey := "testKey1"
	ctx := context.Background()

	unusedRepo.EXPECT().LoadAndDelete(ctx).Return(testKey, nil)
	usedRepo.EXPECT().Store(ctx, testKey, gomock.Any()).Return(true, nil)
	keys := New(KeysConfig{TTL: time.Hour}, usedRepo, unusedRepo)

	actual, err := keys.Use(ctx)
	assert.Equal(t, testKey, actual.Val)
	assert.NotZero(t, actual.ExpiresAt)
	assert.NoError(t, err)
}

func TestKeys_Use_LoadAndDeleteUnexpectedErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usedRepo := NewMockUsedKeysRepository(ctrl)
	unusedRepo := NewMockUnusedKeysRepository(ctrl)
	ctx := context.Background()
	testErr := errors.New("test err")

	unusedRepo.EXPECT().LoadAndDelete(ctx).Return("", testErr)
	keys := New(KeysConfig{TTL: time.Hour}, usedRepo, unusedRepo)

	actual, err := keys.Use(ctx)
	assert.Zero(t, actual)
	assert.EqualError(t, err, testErr.Error())
}

func TestKeys_Use_StoreUnexpectedErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usedRepo := NewMockUsedKeysRepository(ctrl)
	unusedRepo := NewMockUnusedKeysRepository(ctrl)
	testKey := "testKey1"
	ctx := context.Background()
	testErr := errors.New("test err")

	unusedRepo.EXPECT().LoadAndDelete(ctx).Return(testKey, nil)
	usedRepo.EXPECT().Store(ctx, testKey, gomock.Any()).Return(false, testErr)
	keys := New(KeysConfig{TTL: time.Hour}, usedRepo, unusedRepo)

	actual, err := keys.Use(ctx)
	assert.Zero(t, actual)
	assert.EqualError(t, err, testErr.Error())
}

func TestKeys_Use_NoFreeKeys_Retry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usedRepo := NewMockUsedKeysRepository(ctrl)
	unusedRepo := NewMockUnusedKeysRepository(ctrl)
	testKey := "testKey1"
	ctx := context.Background()

	unusedRepo.EXPECT().LoadAndDelete(ctx).Return("", ErrKeyNotFound)
	unusedRepo.EXPECT().LoadAndDelete(ctx).Return(testKey, nil)
	usedRepo.EXPECT().Store(ctx, testKey, gomock.Any()).Return(true, nil)
	keys := New(KeysConfig{TTL: time.Hour}, usedRepo, unusedRepo)

	actual, err := keys.Use(ctx)
	assert.Equal(t, testKey, actual.Val)
	assert.NotZero(t, actual.ExpiresAt)
	assert.NoError(t, err)
}

func TestKeys_Use_KeyAlreadyUsed_Retry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usedRepo := NewMockUsedKeysRepository(ctrl)
	unusedRepo := NewMockUnusedKeysRepository(ctrl)
	testKey1 := "testKey1"
	testKey2 := "testKey2"
	ctx := context.Background()

	unusedRepo.EXPECT().LoadAndDelete(ctx).Return(testKey1, nil)
	unusedRepo.EXPECT().LoadAndDelete(ctx).Return(testKey2, nil)
	usedRepo.EXPECT().Store(ctx, testKey1, gomock.Any()).Return(false, nil)
	usedRepo.EXPECT().Store(ctx, testKey2, gomock.Any()).Return(true, nil)
	keys := New(KeysConfig{TTL: time.Hour}, usedRepo, unusedRepo)

	actual, err := keys.Use(ctx)
	assert.Equal(t, testKey2, actual.Val)
	assert.NotZero(t, actual.ExpiresAt)
	assert.NoError(t, err)
}

func TestKeys_Use_CancelCtx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usedRepo := NewMockUsedKeysRepository(ctrl)
	unusedRepo := NewMockUnusedKeysRepository(ctrl)
	ctx, cancel := context.WithCancel(context.Background())

	unusedRepo.EXPECT().LoadAndDelete(ctx).DoAndReturn(func(context.Context) (string, error) {
		cancel()
		return "", ErrKeyNotFound
	})
	keys := New(KeysConfig{TTL: time.Hour}, usedRepo, unusedRepo)

	actual, err := keys.Use(ctx)
	assert.Zero(t, actual)
	assert.EqualError(t, err, context.Canceled.Error())
}
