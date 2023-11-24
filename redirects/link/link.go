package link

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/demeero/bricks/errbrick"
	"github.com/demeero/bricks/slogbrick"
	linkpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/link/v1beta1"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
		slogbrick.FromCtx(ctx).Error("failed get link from LRU cache",
			slog.String("shortened", shortened),
			slog.String("original", original),
			slog.Any("err", err))
	}
	res, err := l.lookupFromLinkService(ctx, shortened)
	if err != nil {
		return "", err
	}
	go func() {
		rdsCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
		defer cancel()
		err := l.rds.SetArgs(rdsCtx, shortened, res.GetOriginal(), redis.SetArgs{ExpireAt: res.GetExpireTime().AsTime()}).Err()
		if err != nil {
			slogbrick.FromCtx(rdsCtx).Error("failed put link to LRU cache",
				slog.String("shortened", shortened),
				slog.String("original", res.GetOriginal()),
				slog.Any("err", err))
		}
	}()
	return res.GetOriginal(), nil
}

func (l *Links) lookupFromLinkService(ctx context.Context, shortened string) (*linkpb.Link, error) {
	resp, err := l.client.GetLink(ctx, &linkpb.GetLinkRequest{Shortened: shortened})
	if status.Code(err) == codes.NotFound {
		return nil, fmt.Errorf("%w: %s", errbrick.ErrNotFound, shortened)
	}
	if err != nil {
		return nil, err
	}
	return resp.GetLink(), err
}
