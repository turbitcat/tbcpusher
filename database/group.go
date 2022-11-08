package database

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type group struct {
	ID   primitive.ObjectID
	Data any
	db   *MongoDatabase
}

type groupBson struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Data bson.M             `bson:"data,omitempty"`
}

func (g groupBson) toGroup(db *MongoDatabase) group {
	return group{ID: g.ID, Data: g.Data["Value"], db: db}
}

func (g *group) GetID() string {
	return g.ID.Hex()
}

func (g *group) GetData() any {
	return g.Data
}

func (g *group) SetData(data any) error {
	if err := setSomethingById(g.db.ctx, g.db.groupCollection, g.ID, "data", data); err != nil {
		return fmt.Errorf("group setData: %v", err)
	}
	return nil
}

func (g *group) NewSession(hook string, data any) (string, error) {
	db := g.db
	s := sessionBson{Group: g.ID, Hook: hook, Data: bson.M{"Value": data}}
	r, err := db.sessionCollection.InsertOne(db.ctx, s)
	if err != nil {
		return "", fmt.Errorf("newSession: %v", err)
	}
	id := (r.InsertedID).(primitive.ObjectID)
	return id.Hex(), nil
}

func (g *group) GetSessions() ([]Session, error) {
	db := g.db
	cur, err := g.db.sessionCollection.Find(db.ctx, bson.M{"group": g.ID})
	if err != nil {
		return nil, fmt.Errorf("getSessions Find: %v", err)
	}
	var l []sessionBson
	if err = cur.All(db.ctx, &l); err != nil {
		return nil, fmt.Errorf("getSessions All: %v", err)
	}
	l = Filter(l, func(s sessionBson) bool { return !s.Hide })
	f := func(s sessionBson) Session { r := s.toSession(db); return &r }
	return Map(l, f), nil
}
