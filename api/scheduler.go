package api

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/turbitcat/tbcpusher/v2/database"
	"github.com/turbitcat/tbcpusher/v2/scheduler"
	"go.mongodb.org/mongo-driver/bson"
)

func ScheduleGetter(t string, m bson.M) (database.Schedule, error) {
	switch t {
	case "OneTimeSchedule":
		s := OneTimeSchedule{}
		err := s.Load(m)
		return &s, err
	}
	return nil, fmt.Errorf("unknown schedule type: %s", t)
}

func JobGetter(t string, m bson.M) (database.Job, error) {
	switch t {
	case "PushToSessionJob":
		j := PushToSessionJob{}
		err := j.Load(m)
		return &j, err
	}
	return nil, fmt.Errorf("unknown job type: %s", t)
}

type OneTimeSchedule struct {
	scheduler.OneTimeSchedule
}

func NewOneTimeSchedule(t time.Time) *OneTimeSchedule {
	r := OneTimeSchedule{scheduler.OneTimeSchedule{T: t}}
	return &r
}

func (s *OneTimeSchedule) Save() (bson.M, error) {
	t := strconv.FormatInt(s.T.UnixNano(), 16)
	if s.T.IsZero() {
		t = "zero"
	}
	return bson.M{"time": t}, nil
}

func (s *OneTimeSchedule) Load(m bson.M) error {
	_t, ok := m["time"]
	if !ok {
		return fmt.Errorf("missing time")
	}
	t_s, ok := _t.(string)
	if !ok {
		return fmt.Errorf("'time' is not a string")
	}
	if t_s == "zero" {
		s.T = time.Time{}
		return nil
	}
	t, err := strconv.ParseInt(t_s, 16, 64)
	if err != nil {
		return fmt.Errorf("invalid 'time': %v", err)
	}
	s.T = time.Unix(0, t)
	return nil
}

func (s *OneTimeSchedule) GetType() string {
	return "OneTimeSchedule"
}

func (s *OneTimeSchedule) IsType(t string) bool {
	return t == "OneTimeSchedule"
}

type PushToSessionJob struct {
	url  string
	data []byte
}

func NewPushToSessionJob(url string, data []byte) *PushToSessionJob {
	return &PushToSessionJob{url: url, data: data}
}

func (j *PushToSessionJob) Run() {
	http.Post(j.url, "application/json", bytes.NewBuffer(j.data))
}

func (j *PushToSessionJob) Save() (bson.M, error) {
	return bson.M{"url": j.url, "data": string(j.data)}, nil
}

func (j *PushToSessionJob) Load(m bson.M) error {
	url, ok := m["url"]
	if !ok {
		return fmt.Errorf("missing url")
	}
	j.url, ok = url.(string)
	if !ok {
		return fmt.Errorf("url is not a string")
	}
	data, ok := m["data"]
	if !ok {
		return fmt.Errorf("missing data")
	}
	s, ok := data.(string)
	if !ok {
		return fmt.Errorf("data is not a string")
	}
	j.data = []byte(s)
	return nil
}

func (j *PushToSessionJob) GetType() string {
	return "PushToSessionJob"
}

func (j *PushToSessionJob) IsType(t string) bool {
	return t == "PushToSessionJob"
}

type SchedulerLogger struct {
	entries map[scheduler.EntryID]database.Entry
	db      database.Database
}

func NewSchedulerLogger(db database.Database, entries []database.Entry) *SchedulerLogger {
	r := &SchedulerLogger{entries: make(map[scheduler.EntryID]database.Entry), db: db}
	for _, e := range entries {
		r.entries[e.GetID()] = e
	}
	return r
}

func (l *SchedulerLogger) Info(msg string, v ...any) {
	fmt.Println("[Scheduler] ", msg, v)
}

func (l *SchedulerLogger) Error(err error, msg string, v ...any) {
	fmt.Println("[Scheduler] ", err.Error(), msg, v)
}

func (l *SchedulerLogger) EntryAdded(e *scheduler.Entry) {
	entry := database.EntryToDatabaseEntry(l.db, e)
	err := entry.Save()
	if err != nil {
		l.Error(err, "failed to save entry")
	}
	l.entries[e.ID] = entry
}

func (l *SchedulerLogger) EntryRemoved(e *scheduler.Entry) {
	err := l.entries[e.ID].Delete()
	if err != nil {
		l.Error(err, "failed to delete entry")
	}
	delete(l.entries, e.ID)
}

func (l *SchedulerLogger) EntryUpdated(e *scheduler.Entry) {
	entry := database.EntryToDatabaseEntry(l.db, e)
	err := entry.Save()
	if err != nil {
		l.Error(err, "failed to save entry")
	}
	l.entries[e.ID] = entry
}
