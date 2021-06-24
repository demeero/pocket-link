package link

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/demeero/pocket-link/bricks/zaplogger"
	linkpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/link/v1beta1"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrNotFound = errors.New("not found")

type Links struct {
	client linkpb.LinkServiceClient
	rds    redis.Cmdable
}

func New(client linkpb.LinkServiceClient, rds redis.Cmdable) *Links {
	return &Links{
		client: client,
		rds:    rds,
	}
}

func (l *Links) Lookup(ctx context.Context, shortened string) (*url.URL, error) {
	original, err := l.lookup(ctx, shortened)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(original)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	go func() {
		if err := l.rds.Set(ctx, shortened, original, 0).Err(); err != nil {
			zaplogger.From(ctx).Error("error put shortened link to LRU cache",
				zap.String("shortened", shortened), zap.String("original", original))
		}
	}()
	return u, nil
}

func (l *Links) lookup(ctx context.Context, shortened string) (string, error) {
	original, err := l.rds.Get(ctx, shortened).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			zaplogger.From(ctx).Error("error get link from LRU cache",
				zap.String("shortened", shortened), zap.String("original", original), zap.Error(err))
		}
		original, err = l.lookupFromLinkService(ctx, shortened)
	}
	return original, err
}

func (l *Links) lookupFromLinkService(ctx context.Context, shortened string) (string, error) {
	resp, err := l.client.GetLink(ctx, &linkpb.GetLinkRequest{Shortened: shortened})
	if status.Code(err) == codes.NotFound {
		return "", fmt.Errorf("%w: %s", ErrNotFound, shortened)
	}
	if err != nil {
		return "", err
	}
	return resp.GetLink().GetOriginal(), err
}
