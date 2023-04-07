package rpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/keygen/v1beta1"

	"github.com/demeero/pocket-link/keygen/key"
)

type Controller struct {
	pb.KeygenServiceServer
	k *key.Keys
}

func New(k *key.Keys) *Controller {
	return &Controller{k: k}
}

func (c *Controller) GenerateKey(ctx context.Context, _ *pb.GenerateKeyRequest) (*pb.GenerateKeyResponse, error) {
	result, err := c.k.Use(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.GenerateKeyResponse{Key: &pb.Key{
		Val:        result.Val,
		ExpireTime: timestamppb.New(result.ExpiresAt),
	}}, nil
}
