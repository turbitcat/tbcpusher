package database

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Group interface {
	GetID() string
	GetInfo() string
	SetInfo(info string) error
	NewSession(hook string, info string) (string, error)
	GetSessions() ([]Session, error)
}

type Session interface {
	GetID() string
	GetInfo() string
	SetInfo(info string) error
	GetGroup() string
	SetGroup(groupID string) error
	GetPushHook() string
	SetPushHook(url string) error
}

type Database interface {
	NewGroup(info string) (string, error)
	GetGroupByID(id string) (Group, error)
	Close()
}

type group struct {
	ID   primitive.ObjectID
	Info string
	db   *MongoDatabase
}

type session struct {
	ID       primitive.ObjectID
	Group    primitive.ObjectID
	Info     string
	PushHook string
	db       *MongoDatabase
}
