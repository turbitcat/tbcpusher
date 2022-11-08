package database

import (
	"fmt"
	"time"

	"github.com/turbitcat/tbcpusher/v2/scheduler"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	id       primitive.ObjectID
	schedule Schedule
	next     time.Time
	job      Job
	db       *MongoDatabase
}

type entryBson struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Schedule     bson.M             `bson:"schedule,omitempty"`
	ScheduleType string             `bson:"scheduleType,omitempty"`
	Next         time.Time          `bson:"next,omitempty"`
	Prev         time.Time          `bson:"prev,omitempty"`
	Job          bson.M             `bson:"job,omitempty"`
	JobType      string             `bson:"jobType,omitempty"`
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
		id:       e.ID,
		schedule: schedule,
		next:     e.Next,
		job:      job,
	}, nil
}

type scheduleWrapper struct {
	s scheduler.Schedule
	e *entry
}

func (w scheduleWrapper) Next(time.Time) time.Time {
	r := w.s.Next(w.e.next)
	w.e.Save()
	return r
}

func (e *entry) Schedule() scheduler.Schedule {
	return scheduleWrapper{e.schedule, e}
}

func (e *entry) Next() time.Time {
	return e.next
}

func (e *entry) SetNext(next time.Time) {
	e.next = next
	setSomethingById(e.db.ctx, e.db.scheduleCollection, e.id, "next", next)
}

type jobWrapper struct {
	j scheduler.Job
	e *entry
}

func (w jobWrapper) Run() {
	w.j.Run()
	w.e.Save()
}

func (e *entry) Job() scheduler.Job {
	return jobWrapper{e.job, e}
}

func (e *entry) toBson() (entryBson, error) {
	var schedule, job primitive.M
	var scheduleType, jobType string
	var err error
	if e.schedule != nil {
		scheduleType = e.schedule.GetType()
		schedule, err = e.schedule.Save()
		if err != nil {
			return entryBson{}, fmt.Errorf("schedule.Save: %v", err)
		}
	}
	if e.job != nil {
		jobType = e.job.GetType()
		job, err = e.job.Save()
		if err != nil {
			return entryBson{}, fmt.Errorf("job.Save: %v", err)
		}
	}
	return entryBson{
		ID:           e.id,
		Schedule:     schedule,
		ScheduleType: scheduleType,
		Next:         e.next,
		Job:          job,
		JobType:      jobType,
	}, nil
}

func (e *entry) GetID() string {
	return e.id.Hex()
}

func (e *entry) Save() error {
	b, err := e.toBson()
	if err != nil {
		return err
	}
	if e.id.IsZero() {
		r, err := e.db.scheduleCollection.InsertOne(e.db.ctx, b)
		if err != nil {
			return err
		}
		e.id = r.InsertedID.(primitive.ObjectID)
	} else {
		_, err := e.db.scheduleCollection.UpdateOne(e.db.ctx, bson.M{"_id": e.id}, bson.M{"$set": b})
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *entry) Delete() error {
	r, err := e.db.scheduleCollection.DeleteOne(e.db.ctx, bson.M{"_id": e.id})
	if err != nil {
		return err
	}
	if r.DeletedCount != 1 {
		return fmt.Errorf("deleted %d entries", r.DeletedCount)
	}
	return nil
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

func (db *MongoDatabase) GetEntryByID(id string, scheduleGetter SaveableGetter[Schedule], jobGetter SaveableGetter[Job]) (Entry, error) {
	var b entryBson
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %v", err)
	}
	if err := db.groupCollection.FindOne(db.ctx, bson.M{"_id": _id}).Decode(&b); err != nil {
		return nil, err
	}
	e, err := b.toEntry(scheduleGetter, jobGetter)
	if err != nil {
		return nil, err
	}
	e.db = db
	return e, nil
}

type EntryList struct {
	db             Database
	scheduleGetter SaveableGetter[Schedule]
	jobGetter      SaveableGetter[Job]
}

func NewEntryList(db Database, scheduleGetter SaveableGetter[Schedule], jobGetter SaveableGetter[Job]) *EntryList {
	return &EntryList{
		db:             db,
		scheduleGetter: scheduleGetter,
		jobGetter:      jobGetter,
	}
}

func (l *EntryList) NewEntry(job scheduler.Job, schedule scheduler.Schedule) scheduler.Entry {
	return &entry{
		db:       l.db.(*MongoDatabase),
		job:      job.(Job),
		schedule: schedule.(Schedule),
	}
}

func (l *EntryList) Len() int {
	return len(l.All())
}

func (l *EntryList) All() []scheduler.Entry {
	entries, err := l.db.GetAllEntries(l.scheduleGetter, l.jobGetter)
	if err != nil {
		panic(err)
	}
	r := make([]scheduler.Entry, len(entries))
	for i, e := range entries {
		r[i] = e
	}
	return r
}

func (l *EntryList) Add(e scheduler.Entry) {
	if err := e.(Entry).Save(); err != nil {
		fmt.Printf("error saving entry: %v\n", err)
	}
}

func (l *EntryList) Remove(e scheduler.Entry) {
	if err := e.(Entry).Delete(); err != nil {
		fmt.Printf("error deleting entry: %v\n", err)
	}
}
