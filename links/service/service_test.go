package service

import (
	"context"
	"errors"
	"testing"
	"time"

	keygenpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/keygen/v1beta1"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//go:generate mockgen -destination=repo_mock.go -package=service github.com/demeero/pocket-link/links/service Repository
//go:generate mockgen -destination=keygen_client_mock.go -package=service github.com/demeero/pocket-link/proto/gen/go/pocketlink/keygen/v1beta1 KeygenServiceClient

func TestService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	expAt := time.Now().Add(time.Hour)
	createdAt := time.Now()
	orig := "original_test.com"
	short := "shortened_test1"
	createLink := Link{
		Shortened: short,
		Original:  orig,
		ExpAt:     timestamppb.New(expAt).AsTime(),
	}
	expected := Link{
		Shortened: createLink.Shortened,
		Original:  createLink.Original,
		ExpAt:     createLink.ExpAt,
		CreatedAt: createdAt,
	}

	mockRepo := NewMockRepository(ctrl)
	mockKGCli := NewMockKeygenServiceClient(ctrl)
	mockKGCli.EXPECT().GenerateKey(ctx, &keygenpb.GenerateKeyRequest{}).Return(&keygenpb.GenerateKeyResponse{Key: &keygenpb.Key{
		Val:        short,
		ExpireTime: timestamppb.New(expAt),
	}}, nil)
	mockRepo.EXPECT().Create(ctx, createLink).Return(expected, nil)

	svc := New(mockRepo, mockKGCli)

	actual, err := svc.Create(ctx, orig)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestService_Create_InvalidURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	orig := "invalid_url"

	mockRepo := NewMockRepository(ctrl)
	mockKGCli := NewMockKeygenServiceClient(ctrl)

	svc := New(mockRepo, mockKGCli)

	actual, err := svc.Create(ctx, orig)
	assert.Error(t, err)
	assert.Zero(t, actual)
}

func TestService_Create_KeygenClientErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	orig := "original_test.com"

	mockRepo := NewMockRepository(ctrl)
	mockKGCli := NewMockKeygenServiceClient(ctrl)
	testErr := errors.New("test err")
	mockKGCli.EXPECT().GenerateKey(ctx, &keygenpb.GenerateKeyRequest{}).Return(nil, testErr)

	svc := New(mockRepo, mockKGCli)

	actual, err := svc.Create(ctx, orig)
	assert.EqualError(t, err, testErr.Error())
	assert.Zero(t, actual)
}

func TestService_Create_RepoErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	expAt := time.Now().Add(time.Hour)
	orig := "original_test.com"
	short := "shortened_test1"
	createLink := Link{
		Shortened: short,
		Original:  orig,
		ExpAt:     timestamppb.New(expAt).AsTime(),
	}
	testErr := errors.New("test err")

	mockRepo := NewMockRepository(ctrl)
	mockKGCli := NewMockKeygenServiceClient(ctrl)
	mockKGCli.EXPECT().GenerateKey(ctx, &keygenpb.GenerateKeyRequest{}).Return(&keygenpb.GenerateKeyResponse{Key: &keygenpb.Key{
		Val:        short,
		ExpireTime: timestamppb.New(expAt),
	}}, nil)
	mockRepo.EXPECT().Create(ctx, createLink).Return(Link{}, testErr)

	svc := New(mockRepo, mockKGCli)

	actual, err := svc.Create(ctx, orig)
	assert.EqualError(t, err, testErr.Error())
	assert.Zero(t, actual)
}

func TestService_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	expAt := time.Now().Add(time.Hour)
	short := "shortened_test1"

	expected := Link{
		Shortened: short,
		Original:  "original_test.com",
		ExpAt:     timestamppb.New(expAt).AsTime(),
		CreatedAt: time.Now(),
	}

	mockRepo := NewMockRepository(ctrl)
	mockKGCli := NewMockKeygenServiceClient(ctrl)
	mockRepo.EXPECT().LoadByID(ctx, short).Return(expected, nil)

	svc := New(mockRepo, mockKGCli)

	actual, err := svc.Get(ctx, short)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestService_Get_RepoErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	short := "shortened_test1"
	testErr := errors.New("test err")

	mockRepo := NewMockRepository(ctrl)
	mockKGCli := NewMockKeygenServiceClient(ctrl)
	mockRepo.EXPECT().LoadByID(ctx, short).Return(Link{}, testErr)

	svc := New(mockRepo, mockKGCli)

	actual, err := svc.Get(ctx, short)
	assert.EqualError(t, err, testErr.Error())
	assert.Zero(t, actual)
}
