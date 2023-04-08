package link

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	linkpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/link/v1beta1"
	"github.com/rs/zerolog/log"

	"github.com/go-redis/redis/v8"
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
	return u, nil
}

func (l *Links) lookup(ctx context.Context, shortened string) (string, error) {
	original, err := l.rds.Get(ctx, shortened).Result()
	if err == nil {
		return original, nil
	}
	if !errors.Is(err, redis.Nil) {
		log.Ctx(ctx).Error().
			Err(err).
			Str("shortened", shortened).
			Str("original", original).
			Msg("failed get link from LRU cache")
	}
	res, err := l.lookupFromLinkService(ctx, shortened)
	if err != nil {
		return "", err
	}
	go func() {
		err := l.rds.SetArgs(ctx, shortened, res.GetOriginal(), redis.SetArgs{ExpireAt: res.GetExpireTime().AsTime()}).Err()
		if err != nil {
			log.Ctx(ctx).Error().
				Err(err).
				Str("shortened", shortened).
				Str("original", original).
				Msg("failed put link to LRU cache")
		}
	}()
	return res.GetOriginal(), nil
}

func (l *Links) lookupFromLinkService(ctx context.Context, shortened string) (*linkpb.Link, error) {
	resp, err := l.client.GetLink(ctx, &linkpb.GetLinkRequest{Shortened: shortened})
	if status.Code(err) == codes.NotFound {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, shortened)
	}
	if err != nil {
		return nil, err
	}
	return resp.GetLink(), err
}
