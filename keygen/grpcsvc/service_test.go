package grpcsvc

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/keygen/v1beta1"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/demeero/pocket-link/keygen/key"
)

func TestController_GenerateKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usedRepo := key.NewMockUsedKeysRepository(ctrl)
	unusedRepo := key.NewMockUnusedKeysRepository(ctrl)
	testKey := "testKey1"
	ctx := context.Background()

	unusedRepo.EXPECT().LoadAndDelete(ctx).Return(testKey, nil)
	usedRepo.EXPECT().Store(ctx, testKey, gomock.Any()).Return(true, nil)
	keys := key.New(time.Hour, usedRepo, unusedRepo)

	c := New(keys)

	actual, err := c.GenerateKey(ctx, &pb.GenerateKeyRequest{})
	assert.NoError(t, err)
	assert.Equal(t, testKey, actual.GetKey().GetVal())
	assert.NotZero(t, testKey, actual.GetKey().GetExpireTime().AsTime())
}

func TestController_GenerateKey_Err(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usedRepo := key.NewMockUsedKeysRepository(ctrl)
	unusedRepo := key.NewMockUnusedKeysRepository(ctrl)
	testKey := "testKey1"
	ctx := context.Background()

	unusedRepo.EXPECT().LoadAndDelete(ctx).Return(testKey, nil)
	testErr := errors.New("test err")
	usedRepo.EXPECT().Store(ctx, testKey, gomock.Any()).Return(false, testErr)
	keys := key.New(time.Hour, usedRepo, unusedRepo)

	c := New(keys)

	actual, err := c.GenerateKey(ctx, &pb.GenerateKeyRequest{})
	assert.Nil(t, actual)
	assert.ErrorContains(t, err, testErr.Error())
}
