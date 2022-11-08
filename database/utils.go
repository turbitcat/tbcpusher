package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func Map[T any, R any](l []T, f func(T) R) []R {
	r := make([]R, len(l))
	for i, v := range l {
		r[i] = f(v)
	}
	return r
}

func Filter[T any](l []T, f func(T) bool) []T {
	var r []T
	for _, v := range l {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}

func MapErr[T any, R any](l []T, f func(T) (R, error)) ([]R, error) {
	r := make([]R, len(l))
	for i, v := range l {
		m, err := f(v)
		if err != nil {
			return nil, err
		}
		r[i] = m
	}
	return r, nil
}

func setSomethingById(ctx context.Context, collection *mongo.Collection, id any, key string, val any) error {
	r, err := collection.UpdateByID(ctx, id, bson.D{{"$set", bson.D{{key, val}}}})
	if err != nil {
		return err
	}
	if r.MatchedCount != 1 {
		return fmt.Errorf("matched count is %v", r.MatchedCount)
	}
	return nil
}
