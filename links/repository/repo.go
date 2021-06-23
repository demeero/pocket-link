package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/demeero/pocket-link/links/service"
)

type linkMongo struct {
	ID        string    `bson:"_id"`
	Original  string    `bson:"original"`
	CreatedAt time.Time `bson:"created_at"`
	ExpAt     time.Time `bson:"exp_at"`
}

type Repository struct {
	coll *mongo.Collection
}

func New(db *mongo.Database) (*Repository, error) {
	coll := db.Collection("links")
	ind := mongo.IndexModel{
		Keys: bson.M{"exp_at": 1}, Options: options.Index().SetExpireAfterSeconds(0),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if _, err := coll.Indexes().CreateOne(ctx, ind); err != nil {
		return nil, err
	}
	return &Repository{coll: coll}, nil
}

func (r *Repository) Create(ctx context.Context, link service.Link) (service.Link, error) {
	lm := linkMongo{
		ID:        link.Shortened,
		Original:  link.Original,
		CreatedAt: time.Now().UTC(),
		ExpAt:     link.ExpAt,
	}
	_, err := r.coll.InsertOne(ctx, lm)
	if err != nil {
		return service.Link{}, err
	}
	link.CreatedAt = lm.CreatedAt
	return link, nil
}

func (r *Repository) LoadByID(ctx context.Context, shortened string) (service.Link, error) {
	res := r.coll.FindOne(ctx, bson.M{"_id": shortened})
	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return service.Link{}, fmt.Errorf("%w: %s", service.ErrNotFound, shortened)
	}
	lm := linkMongo{}
	if err := res.Decode(&lm); err != nil {
		return service.Link{}, err
	}
	return service.Link{
		Shortened: lm.ID,
		Original:  lm.Original,
		CreatedAt: lm.CreatedAt,
		ExpAt:     lm.ExpAt,
	}, nil
}
