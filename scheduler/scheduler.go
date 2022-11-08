package scheduler

import (
	"sync"
	"time"
)

type Entry interface {
	Schedule() Schedule
	Next() time.Time
	SetNext(time.Time)
	Job() Job
}

type EntryList interface {
	NewEntry(Job, Schedule) Entry
	Len() int
	All() []Entry
	Add(Entry)
	Remove(Entry)
}

type Scheduler struct {
	entries   EntryList
	add       chan Entry
	remove    chan Entry
	stop      chan struct{}
	running   bool
	runningMu sync.Mutex
	location  *time.Location
	logger    Logger
}

func newScheduler() *Scheduler {
	return &Scheduler{add: make(chan Entry), remove: make(chan Entry), stop: make(chan struct{})}
}

func NewDefult() *Scheduler {
	s := newScheduler()
	s.logger = PrintlnLogger()
	s.location = time.Local
	return s
}

func (s *Scheduler) SetLogger(l Logger) {
	if s.running {
		panic("cannot set logger while running")
	}
	s.logger = l
}

func (s *Scheduler) SetEntries(entries EntryList) {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()
	if s.running {
		panic("cannot set entries while running")
	}
	s.entries = entries
}

func (s *Scheduler) AddJob(job Job, schedule Schedule) {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()
	entry := s.entries.NewEntry(job, schedule)
	if s.running {
		s.add <- entry
	} else {
		s.addEntry(entry)
	}
}

func (s *Scheduler) AddFunc(f func(), schedule Schedule) {
	s.AddJob(FuncJob(f), schedule)
}

func (s *Scheduler) Remove(entry Entry) {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()
	if s.running {
		s.remove <- entry
	} else {
		s.removeEntry(entry)
	}
}

func (s *Scheduler) removeEntry(entry Entry) {
	s.entries.Remove(entry)
}

func (s *Scheduler) addEntry(entry Entry) {
	s.entries.Add(entry)
}

func (s *Scheduler) now() time.Time {
	return time.Now().In(s.location)
}

func (s *Scheduler) updateEntries(now time.Time) {

}

func (s *Scheduler) next(now time.Time) time.Time {
	t := time.Time{}
	for _, v := range s.entries.All() {
		next := v.Next()
		if next.IsZero() {
			continue
		}
		if t.IsZero() || next.Before(t) {
			t = next
		}
	}
	return t
}

func (s *Scheduler) Run() {
	s.runningMu.Lock()
	if s.running {
		s.runningMu.Unlock()
		return
	}
	s.running = true
	s.runningMu.Unlock()
	go s.run()
}

func (s *Scheduler) run() {
	now := s.now()
	s.logger.Info("start", "now", now)
	s.updateEntries(now)
	for {
		next := s.next(now)
		s.logger.Info("next", "time", next)
		var timer *time.Timer
		if next.IsZero() {
			timer = time.NewTimer(time.Hour * 240000)
		} else {
			timer = time.NewTimer(next.Sub(now))
		}
		for {
			select {
			case now = <-timer.C:
				now = now.In(s.location)
				s.logger.Info("wake", "now", now)
				for _, v := range s.entries.All() {
					next := v.Next()
					if !next.After(now) {
						startJob(v)
						n := v.Schedule().Next(now)
						v.SetNext(n)
						if n.IsZero() {
							s.removeEntry(v)
						}
						s.logger.Info("jobRunning", "job", v.Job(), "now", now, "next", n)
					}
				}
			case newEntry := <-s.add:
				timer.Stop()
				now = s.now()
				n := newEntry.Schedule().Next(now)
				newEntry.SetNext(n)
				if !n.IsZero() {
					s.addEntry(newEntry)
					s.logger.Info("jobAdded", "now", now, "next", n)
				}
			case id := <-s.remove:
				timer.Stop()
				now = s.now()
				s.removeEntry(id)
				s.logger.Info("jobRemoved", "id", id)
			case <-s.stop:
				timer.Stop()
				s.logger.Info("stop")
				return
			}
			break
		}
	}
}

func startJob(entry Entry) {
	go func() {
		entry.Job().Run()
	}()
}
