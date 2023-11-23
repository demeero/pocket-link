package rpc

import (
	"context"
	"errors"

	"github.com/demeero/bricks/errbrick"
	pb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/link/v1beta1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/demeero/pocket-link/links/service"
)

type Controller struct {
	pb.UnimplementedLinkServiceServer
	svc *service.Service
}

func New(s *service.Service) *Controller {
	return &Controller{svc: s}
}

func (c *Controller) GetLink(ctx context.Context, req *pb.GetLinkRequest) (*pb.GetLinkResponse, error) {
	l, err := c.svc.Get(ctx, req.GetShortened())
	if errors.Is(err, errbrick.ErrNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, err
	}
	return &pb.GetLinkResponse{Link: &pb.Link{
		Original:   l.Original,
		Shortened:  l.Shortened,
		CreateTime: timestamppb.New(l.CreatedAt),
		ExpireTime: timestamppb.New(l.ExpAt),
	}}, nil
}
