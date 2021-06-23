package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"

	keygenpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/keygen/v1beta1"
)

var (
	ErrInvalid  = errors.New("invalid data")
	ErrNotFound = errors.New("not found")
)

type Link struct {
	Shortened string    `json:"shortened,omitempty"`
	Original  string    `json:"original,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	ExpAt     time.Time `json:"exp_at"`
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
		return Link{}, fmt.Errorf("%w: invalid url: %s", ErrInvalid, original)
	}
	resp, err := s.keygenClient.GenerateKey(ctx, &keygenpb.GenerateKeyRequest{})
	if err != nil {
		return Link{}, err
	}
	return s.repo.Create(ctx, Link{
		Shortened: resp.GetKey().GetVal(),
		Original:  original,
		ExpAt:     resp.GetKey().GetExpireTime().AsTime(),
	})
}

func (s *Service) Get(ctx context.Context, shortened string) (Link, error) {
	return s.repo.LoadByID(ctx, shortened)
}
