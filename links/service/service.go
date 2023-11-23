package service

import (
	"context"
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/demeero/bricks/errbrick"

	keygenpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/keygen/v1beta1"
)

type Link struct {
	CreatedAt time.Time `json:"created_at"`
	ExpAt     time.Time `json:"exp_at"`
	Shortened string    `json:"shortened,omitempty"`
	Original  string    `json:"original,omitempty"`
}

type Repository interface {
	Create(context.Context, Link) (Link, error)
	LoadByID(context.Context, string) (Link, error)
}

type Service struct {
	repo         Repository
	keygenClient keygenpb.KeygenServiceClient
}

func New(repo Repository, kc keygenpb.KeygenServiceClient) *Service {
	return &Service{
		repo:         repo,
		keygenClient: kc,
	}
}

func (s *Service) Create(ctx context.Context, original string) (Link, error) {
	if !govalidator.IsURL(original) {
		return Link{}, fmt.Errorf("%w: incorrect url format: %s", errbrick.ErrInvalidData, original)
	}
	resp, err := s.keygenClient.GenerateKey(ctx, &keygenpb.GenerateKeyRequest{})
	if err != nil {
		return Link{}, fmt.Errorf("failed generate key: %w", err)
	}
	link, err := s.repo.Create(ctx, Link{
		Shortened: resp.GetKey().GetVal(),
		Original:  original,
		ExpAt:     resp.GetKey().GetExpireTime().AsTime(),
	})
	if err != nil {
		return Link{}, fmt.Errorf("failed create link: %w", err)
	}
	return link, nil
}

func (s *Service) Get(ctx context.Context, shortened string) (Link, error) {
	link, err := s.repo.LoadByID(ctx, shortened)
	if err != nil {
		return Link{}, fmt.Errorf("failed load link: %w", err)
	}
	return link, nil
}
