package database

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Group interface {
	GetID() string
	GetData() any
	SetData(data any) error
	NewSession(hook string, data any) (string, error)
	GetSessions() ([]Session, error)
}

type Session interface {
	GetID() string
	GetData() any
	SetData(data any) error
	GetGroupID() string
	SetGroupID(groupID string) error
	GetGroup() (Group, error)
	GetPushHook() string
	SetPushHook(url string) error
	Hide() error
}

type Database interface {
	NewGroup(data any) (string, error)
	NewSession(hook string, data any) (string, error)
	GetGroupByID(id string) (Group, error)
	GetSessionByID(id string) (Session, error)
	GetAllGroups() ([]Group, error)
	SaveState(name string, data any) error
	BindState(name string, data any) error
	Close()
}

type group struct {
	ID   primitive.ObjectID
	Data any
	db   *MongoDatabase
}

type session struct {
	ID       primitive.ObjectID
	Group    primitive.ObjectID
	Data     any
	PushHook string
	db       *MongoDatabase
}
