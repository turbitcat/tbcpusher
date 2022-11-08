package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDatabase struct {
	tbcPushDatabase    *mongo.Database
	groupCollection    *mongo.Collection
	sessionCollection  *mongo.Collection
	stateCollection    *mongo.Collection
	scheduleCollection *mongo.Collection
	ctx                context.Context
	client             *mongo.Client
}

func NewMongo(atlasURI string, database string) (Database, error) {

	client, err := mongo.NewClient(options.Client().ApplyURI(atlasURI))
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}
	d := client.Database(database)
	gr := d.Collection("group")
	se := d.Collection("session")
	sc := d.Collection("schedule")
	db := MongoDatabase{
		ctx:                ctx,
		tbcPushDatabase:    d,
		groupCollection:    gr,
		sessionCollection:  se,
		scheduleCollection: sc,
		client:             client,
	}
	return &db, nil
}

func (db *MongoDatabase) Close() {
	db.client.Disconnect(db.ctx)
}

func (db *MongoDatabase) NewGroup(data any) (string, error) {
	g := groupBson{Data: bson.M{"Value": data}}
	r, err := db.groupCollection.InsertOne(db.ctx, g)
	if err != nil {
		return "", fmt.Errorf("newGroup: %v", err)
	}
	id := (r.InsertedID).(primitive.ObjectID)
	return id.Hex(), nil
}

func (db *MongoDatabase) GetGroupByID(id string) (Group, error) {
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("getGroupByID invalid id \"%v\": %v", id, err)
	}
	r := db.groupCollection.FindOne(db.ctx, bson.D{{"_id", _id}})
	g := groupBson{}
	if err := r.Decode(&g); err != nil {
		return nil, fmt.Errorf("getGroupByID: %v", err)
	}
	group := g.toGroup(db)
	return &group, nil
}

func (db *MongoDatabase) GetSessionByID(id string) (Session, error) {
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("getSessionByID invalid id \"%v\": %v", id, err)
	}
	r := db.sessionCollection.FindOne(db.ctx, bson.D{{"_id", _id}})
	g := sessionBson{}
	if err := r.Decode(&g); err != nil {
		return nil, fmt.Errorf("getSessionByID: %v", err)
	}
	session := g.toSession(db)
	return &session, nil
}

func (db *MongoDatabase) GetAllGroups() ([]Group, error) {
	cur, err := db.groupCollection.Find(db.ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("getAllGroup Find: %v", err)
	}
	var l []groupBson
	if err = cur.All(db.ctx, &l); err != nil {
		return nil, fmt.Errorf("getAllgroup All: %v", err)
	}
	f := func(g groupBson) Group { t := g.toGroup(db); return &t }
	return Map(l, f), nil
}

func (db *MongoDatabase) NewSession(hook string, data any) (string, error) {
	s := sessionBson{Hook: hook, Data: bson.M{"Value": data}}
	r, err := db.sessionCollection.InsertOne(db.ctx, s)
	if err != nil {
		return "", fmt.Errorf("newSession: %v", err)
	}
	id := (r.InsertedID).(primitive.ObjectID)
	return id.Hex(), nil
}
