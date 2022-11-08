package database

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type session struct {
	ID       primitive.ObjectID
	Group    primitive.ObjectID
	Data     any
	PushHook string
	db       *MongoDatabase
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
