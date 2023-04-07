package mongo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type key struct {
	ExpAt time.Time `bson:"exp_at"`
	ID    string    `bson:"_id"`
}

type UsedKeys struct {
	coll *mongo.Collection
}

func NewUsedKeys(db *mongo.Database) (*UsedKeys, error) {
	coll := db.Collection("used_keys")
	ind := mongo.IndexModel{
		Keys: bson.M{"exp_at": 1}, Options: options.Index().SetExpireAfterSeconds(0),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if _, err := coll.Indexes().CreateOne(ctx, ind); err != nil {
		return nil, err
	}
	return &UsedKeys{coll: coll}, nil
}

func (u *UsedKeys) Store(ctx context.Context, k string, ttl time.Duration) (bool, error) {
	_, err := u.coll.InsertOne(ctx, key{
		ID:    k,
		ExpAt: time.Now().Add(ttl).UTC(),
	})
	if mongo.IsDuplicateKeyError(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (u *UsedKeys) Exists(ctx context.Context, k string) (bool, error) {
	err := u.coll.FindOne(ctx, bson.M{"_id": k}).Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
