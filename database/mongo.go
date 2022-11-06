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
	tbcPushDatabase   *mongo.Database
	groupCollection   *mongo.Collection
	sessionCollection *mongo.Collection
	stateCollection   *mongo.Collection
	ctx               context.Context
	client            *mongo.Client
}

type groupBson struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Data bson.M             `bson:"data,omitempty"`
}

func (g groupBson) toGroup(db *MongoDatabase) group {
	return group{ID: g.ID, Data: g.Data["Value"], db: db}
}

type sessionBson struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Group primitive.ObjectID `bson:"group,omitempty"`
	Data  bson.M             `bson:"data,omitempty"`
	Hook  string             `bson:"hook,omitempty"`
	Hide  bool               `bson:"hide,omitempty"`
}

func (s sessionBson) toSession(db *MongoDatabase) session {
	return session{ID: s.ID, Group: s.Group, Data: s.Data["Value"], db: db, PushHook: s.Hook}
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
	st := d.Collection("state")
	db := MongoDatabase{
		ctx:               ctx,
		tbcPushDatabase:   d,
		groupCollection:   gr,
		sessionCollection: se,
		client:            client,
		stateCollection:   st,
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

func (db *MongoDatabase) SaveState(name string, data any) error {
	db.stateCollection.FindOneAndUpdate(db.ctx, bson.M{"name": name})
}

func (db *MongoDatabase) BindState(name string, data any) error

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

func (s *session) GetID() string {
	return s.ID.Hex()
}

func (s *session) GetData() any {
	return s.Data
}

func (s *session) SetData(data any) error {
	if err := setSomethingById(s.db.ctx, s.db.sessionCollection, s.ID, "data", data); err != nil {
		return fmt.Errorf("session setData: %v", err)
	}
	return nil
}

func (s *session) GetGroupID() string {
	return s.Group.Hex()
}

func (s *session) SetGroupID(groupID string) error {
	id, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return fmt.Errorf("session setGroup invalid id: %v", err)
	}
	if err := setSomethingById(s.db.ctx, s.db.sessionCollection, s.ID, "group", id); err != nil {
		return fmt.Errorf("session setGroup: %v", err)
	}
	return nil
}

func (s *session) GetGroup() (Group, error) {
	return s.db.GetGroupByID(string(s.Group.Hex()))
}

func (s *session) GetPushHook() string {
	return s.PushHook
}

func (s *session) SetPushHook(url string) error {
	if err := setSomethingById(s.db.ctx, s.db.sessionCollection, s.ID, "hook", url); err != nil {
		return fmt.Errorf("session setPushHook: %v", err)
	}
	return nil
}

func (s *session) Hide() error {
	if err := setSomethingById(s.db.ctx, s.db.sessionCollection, s.ID, "hide", true); err != nil {
		return fmt.Errorf("session hide: %v", err)
	}
	return nil
}
