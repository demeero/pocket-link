package link

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	linkpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/link/v1beta1"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//go:generate mockgen -destination=link_client_mock.go -package=link github.com/demeero/pocket-link/proto/gen/go/pocketlink/link/v1beta1 LinkServiceClient

func TestLinks_Lookup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	short := "shortened_test1"
	orig := "https://original.com"
	expected, err := url.Parse(orig)
	require.NoError(t, err)

	mockLinkClient := NewMockLinkServiceClient(ctrl)
	expAt := timestamppb.New(time.Now().Add(time.Hour))
	mockLinkClient.EXPECT().GetLink(ctx, &linkpb.GetLinkRequest{Shortened: short}).Return(&linkpb.GetLinkResponse{Link: &linkpb.Link{
		Original:   orig,
		Shortened:  short,
		CreateTime: timestamppb.New(time.Now()),
		ExpireTime: expAt,
	}}, nil)

	db, mock := redismock.NewClientMock()
	mock.ExpectGet(short).RedisNil()

	l := New(mockLinkClient, db)

	actual, err := l.Lookup(ctx, short)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLinks_Lookup_FromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	rdsClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	ctx := context.Background()
	short := "shortened_test1"
	orig := "https://original.com"
	expected, err := url.Parse(orig)
	require.NoError(t, err)

	mockLinkClient := NewMockLinkServiceClient(ctrl)

	rdsClient.Set(ctx, short, orig, time.Hour)

	l := New(mockLinkClient, rdsClient)

	actual, err := l.Lookup(ctx, short)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestLinks_Lookup_RedisErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	short := "shortened_test1"
	orig := "https://original.com"
	expected, err := url.Parse(orig)
	require.NoError(t, err)

	mockLinkClient := NewMockLinkServiceClient(ctrl)
	mockLinkClient.EXPECT().GetLink(ctx, &linkpb.GetLinkRequest{Shortened: short}).Return(&linkpb.GetLinkResponse{Link: &linkpb.Link{
		Original:   orig,
		Shortened:  short,
		CreateTime: timestamppb.New(time.Now()),
		ExpireTime: timestamppb.New(time.Now().Add(time.Hour)),
	}}, nil)

	db, mock := redismock.NewClientMock()
	mock.ExpectGet(short).SetErr(redis.ErrClosed)

	l := New(mockLinkClient, db)

	actual, err := l.Lookup(ctx, short)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestLinks_Lookup_LinkServiceErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	short := "shortened_test1"
	testErr := errors.New("test err")

	mockLinkClient := NewMockLinkServiceClient(ctrl)
	mockLinkClient.EXPECT().GetLink(ctx, &linkpb.GetLinkRequest{Shortened: short}).Return(nil, testErr)

	db, mock := redismock.NewClientMock()
	mock.ExpectGet(short).RedisNil()

	l := New(mockLinkClient, db)

	actual, err := l.Lookup(ctx, short)
	assert.EqualError(t, err, testErr.Error())
	assert.Nil(t, actual)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLinks_Lookup_WithoutSchema(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	rdsClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	ctx := context.Background()
	short := "shortened_test1"
	orig := "original.com"
	expected, err := url.Parse("http://" + orig)
	require.NoError(t, err)

	mockLinkClient := NewMockLinkServiceClient(ctrl)

	rdsClient.Set(ctx, short, orig, time.Hour)

	l := New(mockLinkClient, rdsClient)

	actual, err := l.Lookup(ctx, short)
	assert.NoError(t, err)
	assert.Equal(t, expected.String(), actual.String())
}

func TestLinks_Lookup_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	short := "shortened_test1"

	mockLinkClient := NewMockLinkServiceClient(ctrl)

	mockLinkClient.EXPECT().GetLink(ctx, &linkpb.GetLinkRequest{Shortened: short}).
		Return(nil, status.Error(codes.NotFound, "not found"))

	db, mock := redismock.NewClientMock()
	mock.ExpectGet(short).RedisNil()

	l := New(mockLinkClient, db)

	actual, err := l.Lookup(ctx, short)
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, actual)
	assert.NoError(t, mock.ExpectationsWereMet())
}
