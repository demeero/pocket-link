package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	keygenpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/keygen/v1beta1"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/demeero/pocket-link/links/service"
)

func Test_create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	expAt := time.Now().Add(time.Hour)
	createdAt := time.Now()
	orig := "original_test.com"
	short := "shortened_test1"
	l := service.Link{
		Shortened: short,
		Original:  orig,
		ExpAt:     timestamppb.New(expAt).AsTime(),
	}
	expected := service.Link{
		Shortened: l.Shortened,
		Original:  l.Original,
		ExpAt:     l.ExpAt,
		CreatedAt: timestamppb.New(createdAt).AsTime(),
	}

	mockRepo := service.NewMockRepository(ctrl)
	mockKGCli := service.NewMockKeygenServiceClient(ctrl)
	mockKGCli.EXPECT().GenerateKey(ctx, &keygenpb.GenerateKeyRequest{}).Return(&keygenpb.GenerateKeyResponse{Key: &keygenpb.Key{
		Val:        short,
		ExpireTime: timestamppb.New(expAt),
	}}, nil)
	mockRepo.EXPECT().Create(ctx, l).Return(expected, nil)

	cl := createLink{Original: orig}
	b, err := json.Marshal(cl)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	err = create(service.New(mockRepo, mockKGCli))(c)
	actual := service.Link{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &actual))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, echo.MIMEApplicationJSONCharsetUTF8, rec.Header().Get(echo.HeaderContentType))
	assert.Equal(t, expected, actual)
}

func Test_create_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := service.NewMockRepository(ctrl)
	mockKGCli := service.NewMockKeygenServiceClient(ctrl)

	cl := createLink{Original: "blabla_url"}
	b, err := json.Marshal(cl)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	err = create(service.New(mockRepo, mockKGCli))(c)
	assert.Equal(t, err, echo.NewHTTPError(http.StatusBadRequest, "invalid data: invalid url: blabla_url"))
}
