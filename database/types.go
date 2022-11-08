package database

import "github.com/turbitcat/tbcpusher/v2/scheduler"

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
	GetAllEntries(SaveableGetter[Schedule], SaveableGetter[Job]) ([]Entry, error)
	GetMaxEntryID() (scheduler.EntryID, error)
	GetEntryByID(scheduler.EntryID, SaveableGetter[Schedule], SaveableGetter[Job]) (Entry, error)
	Close()
}

type Entry interface {
	Save() error
	Delete() error
	ToEntry() scheduler.Entry
	GetID() scheduler.EntryID
}
