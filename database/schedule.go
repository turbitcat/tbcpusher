package database

import (
	"fmt"
	"time"

	"github.com/turbitcat/tbcpusher/v2/scheduler"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Saveable interface {
	GetType() string
	Save() (bson.M, error)
	Load(bson.M) error
	IsType(string) bool
}

type Schedule interface {
	scheduler.Schedule
	Saveable
}

type Job interface {
	scheduler.Job
	Saveable
}

type entry struct {
	ID       scheduler.EntryID
	Schedule Schedule
	Next     time.Time
	Prev     time.Time
	Job      Job
	db       *MongoDatabase
}

type entryBson struct {
	ID           scheduler.EntryID `bson:"_id,omitempty"`
	Schedule     bson.M            `bson:"schedule,omitempty"`
	ScheduleType string            `bson:"scheduleType,omitempty"`
	Next         time.Time         `bson:"next,omitempty"`
	Prev         time.Time         `bson:"prev,omitempty"`
	Job          bson.M            `bson:"job,omitempty"`
	JobType      string            `bson:"jobType,omitempty"`
}

type SaveableGetter[T Saveable] func(string, bson.M) (T, error)

func (e entryBson) toEntry(scheduleGetter SaveableGetter[Schedule], jobGetter SaveableGetter[Job]) (*entry, error) {
	schedule, err := scheduleGetter(e.ScheduleType, e.Schedule)
	if err != nil {
		return nil, fmt.Errorf("scheduleGetter: %v", err)
	}
	job, err := jobGetter(e.JobType, e.Job)
	if err != nil {
		return nil, fmt.Errorf("jobGetter: %v", err)
	}
	return &entry{
		ID:       e.ID,
		Schedule: schedule,
		Next:     e.Next,
		Prev:     e.Prev,
		Job:      job,
	}, nil
}

func EntryToDatabaseEntry(db Database, e *scheduler.Entry) Entry {
	return &entry{
		ID:       e.ID,
		Schedule: e.Schedule.(Schedule),
		Next:     e.Next,
		Prev:     e.Prev,
		Job:      e.Job.(Job),
		db:       db.(*MongoDatabase),
	}
}

func SetUpScheduler(db Database, s *scheduler.Scheduler, scheduleGetter SaveableGetter[Schedule], jobGetter SaveableGetter[Job]) []Entry {
	entries, err := db.GetAllEntries(scheduleGetter, jobGetter)
	if err == nil {
		etr := []scheduler.Entry{}
		for _, e := range entries {
			etr = append(etr, e.ToEntry())
		}
		s.SetEntries(etr)
	} else {
		fmt.Printf("error getting entries: %v\n", err)
	}
	maxId, err := db.GetMaxEntryID()
	if err != nil {
		maxId = 0
	}
	s.SetNextID(maxId + 1)
	return entries
}

func (e *entry) toBson() (entryBson, error) {
	schedule, err := e.Schedule.Save()
	if err != nil {
		return entryBson{}, err
	}
	job, err := e.Job.Save()
	if err != nil {
		return entryBson{}, err
	}
	return entryBson{
		ID:           e.ID,
		Schedule:     schedule,
		ScheduleType: e.Schedule.GetType(),
		Next:         e.Next,
		Prev:         e.Prev,
		Job:          job,
		JobType:      e.Job.GetType(),
	}, nil
}

func (e *entry) Save() error {
	b, err := e.toBson()
	if err != nil {
		return err
	}
	if e.ID == 0 {
		return fmt.Errorf("entry id is 0")
	} else {
		if _, err := e.db.scheduleCollection.ReplaceOne(e.db.ctx, bson.M{"_id": e.ID}, b, options.Replace().SetUpsert(true)); err != nil {
			return err
		}
	}
	return nil
}

func (e *entry) Delete() error {
	if e.ID == 0 {
		return nil
	}
	if _, err := e.db.scheduleCollection.DeleteOne(e.db.ctx, bson.M{"_id": e.ID}); err != nil {
		return err
	}
	return nil
}

func (e *entry) ToEntry() scheduler.Entry {
	return scheduler.Entry{
		ID:       e.ID,
		Schedule: e.Schedule,
		Next:     e.Next,
		Prev:     e.Prev,
		Job:      e.Job,
	}
}

func (e *entry) GetID() scheduler.EntryID {
	return e.ID
}

func (db *MongoDatabase) GetAllEntries(scheduleGetter SaveableGetter[Schedule], jobGetter SaveableGetter[Job]) ([]Entry, error) {
	cur, err := db.scheduleCollection.Find(db.ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var entries []Entry
	for cur.Next(db.ctx) {
		var b entryBson
		if err := cur.Decode(&b); err != nil {
			return nil, err
		}
		e, err := b.toEntry(scheduleGetter, jobGetter)
		if err != nil {
			return nil, err
		}
		e.db = db
		entries = append(entries, e)
	}
	return entries, nil
}

func (db *MongoDatabase) GetEntryByID(id scheduler.EntryID, scheduleGetter SaveableGetter[Schedule], jobGetter SaveableGetter[Job]) (Entry, error) {
	var b entryBson
	if err := db.groupCollection.FindOne(db.ctx, bson.M{"_id": id}).Decode(&b); err != nil {
		return nil, err
	}
	e, err := b.toEntry(scheduleGetter, jobGetter)
	if err != nil {
		return nil, err
	}
	e.db = db
	return e, nil
}

func (db *MongoDatabase) GetMaxEntryID() (scheduler.EntryID, error) {
	var b entryBson
	if err := db.scheduleCollection.FindOne(db.ctx, bson.M{}, options.FindOne().SetSort(bson.D{{"_id", -1}})).Decode(&b); err != nil {
		return 0, err
	}
	return b.ID, nil
}
