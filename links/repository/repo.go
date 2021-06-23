package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/demeero/pocket-link/links/service"
)

type createLink struct {
	Shortened string    `bson:"_id"`
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
	if _, err := coll.Indexes().CreateOne(context.TODO(), ind); err != nil {
		return nil, err
	}
	return &Repository{coll: coll}, nil
}

func (r *Repository) Create(ctx context.Context, link service.Link) (service.Link, error) {
	cl := createLink{
		Shortened: link.Shortened,
		Original:  link.Original,
		CreatedAt: time.Now().UTC(),
		ExpAt:     link.ExpAt,
	}
	_, err := r.coll.InsertOne(ctx, cl)
	if err != nil {
		return service.Link{}, err
	}
	link.CreatedAt = cl.CreatedAt
	return link, nil
}
