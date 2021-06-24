package rpc

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/link/v1beta1"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/demeero/pocket-link/links/service"
)

func TestController_GetLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	short := "shortened_test1"
	expAt := time.Now().Add(time.Hour)

	expected := &pb.GetLinkResponse{Link: &pb.Link{
		Original:   "original.com",
		Shortened:  short,
		CreateTime: timestamppb.New(time.Now()),
		ExpireTime: timestamppb.New(expAt),
	}}

	mockRepo := service.NewMockRepository(ctrl)
	mockKGCli := service.NewMockKeygenServiceClient(ctrl)
	mockRepo.EXPECT().LoadByID(ctx, short).Return(service.Link{
		Shortened: expected.GetLink().GetShortened(),
		Original:  expected.GetLink().GetOriginal(),
		CreatedAt: expected.GetLink().GetCreateTime().AsTime(),
		ExpAt:     expected.GetLink().GetExpireTime().AsTime(),
	}, nil)

	c := New(service.New(mockRepo, mockKGCli))

	actual, err := c.GetLink(ctx, &pb.GetLinkRequest{Shortened: short})
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestController_GetLink_ErrNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	short := "shortened_test1"

	mockRepo := service.NewMockRepository(ctrl)
	mockKGCli := service.NewMockKeygenServiceClient(ctrl)
	mockRepo.EXPECT().LoadByID(ctx, short).Return(service.Link{}, service.ErrNotFound)

	c := New(service.New(mockRepo, mockKGCli))

	actual, err := c.GetLink(ctx, &pb.GetLinkRequest{Shortened: short})

	assert.Equal(t, codes.NotFound, status.Code(err))
	assert.Nil(t, actual)
}

func TestController_GetLink_UnexpectedErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	short := "shortened_test1"
	testErr := errors.New("test err")

	mockRepo := service.NewMockRepository(ctrl)
	mockKGCli := service.NewMockKeygenServiceClient(ctrl)
	mockRepo.EXPECT().LoadByID(ctx, short).Return(service.Link{}, testErr)

	c := New(service.New(mockRepo, mockKGCli))

	actual, err := c.GetLink(ctx, &pb.GetLinkRequest{Shortened: short})
	assert.Error(t, err)
	assert.Nil(t, actual)
}
