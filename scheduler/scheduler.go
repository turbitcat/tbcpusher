package scheduler

import (
	"sync"
	"time"
)

type EntryID int

type Entry struct {
	ID       EntryID
	Schedule Schedule
	Next     time.Time
	Prev     time.Time
	Job      Job
}

type Scheduler struct {
	entries   []*Entry
	add       chan *Entry
	remove    chan EntryID
	stop      chan struct{}
	snap      chan chan []Entry
	running   bool
	runningMu sync.Mutex
	nextID    EntryID
	location  *time.Location
	logger    Logger
}

func newScheduler() *Scheduler {
	return &Scheduler{add: make(chan *Entry), remove: make(chan EntryID), stop: make(chan struct{}), snap: make(chan chan []Entry)}
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

func (s *Scheduler) SetNextID(id EntryID) {
	if s.running {
		panic("scheduler is running")
	}
	s.nextID = id
}

func (s *Scheduler) SetEntries(entries []Entry) {
	if s.running {
		panic("cannot set entries while running")
	}
	s.runningMu.Lock()
	defer s.runningMu.Unlock()
	s.entries = make([]*Entry, len(entries))
	for i, e := range entries {
		s.entries[i] = &e
	}
}

func (s *Scheduler) AddJob(job Job, schedule Schedule) EntryID {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()
	entry := &Entry{
		ID:       s.nextID,
		Schedule: schedule,
		Job:      job,
	}
	s.nextID++
	if s.running {
		s.add <- entry
	} else {
		s.addEntry(entry)
	}
	return entry.ID
}

func (s *Scheduler) AddFunc(f func(), schedule Schedule) EntryID {
	return s.AddJob(FuncJob(f), schedule)
}

func (s *Scheduler) GetEntries() []Entry {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()
	if s.running {
		c := make(chan []Entry, 1)
		s.snap <- c
		return <-c
	}
	return s.getEntries()
}

func (s *Scheduler) getEntries() []Entry {
	entries := make([]Entry, len(s.entries))
	for i, v := range s.entries {
		entries[i] = *v
	}
	return entries
}

func (s *Scheduler) Remove(id EntryID) {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()
	if s.running {
		s.remove <- id
	} else {
		s.removeEntry(id)
	}
}

func (s *Scheduler) removeEntry(id EntryID) {
	entries := []*Entry{}
	for _, v := range s.entries {
		if v.ID != id {
			entries = append(entries, v)
		} else {
			s.logger.EntryRemoved(v)
		}
	}
	s.entries = entries
}

func (s *Scheduler) addEntry(entry *Entry) {
	s.entries = append(s.entries, entry)
	s.logger.EntryAdded(entry)
}

func (s *Scheduler) now() time.Time {
	return time.Now().In(s.location)
}

func (s *Scheduler) updateEntries(now time.Time) {
	entries := []*Entry{}
	for _, v := range s.entries {
		if !v.Next.IsZero() {
			entries = append(entries, v)
			continue
		}
		v.Next = v.Schedule.Next(now)
		if !v.Next.IsZero() {
			s.logger.EntryAdded(v)
			entries = append(entries, v)
		}
	}
	s.entries = entries
}

func (s *Scheduler) next(now time.Time) time.Time {
	t := time.Time{}
	for _, v := range s.entries {
		if v.Next.IsZero() {
			continue
		}
		if t.IsZero() || v.Next.Before(t) {
			t = v.Next
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
				for _, v := range s.entries {
					if !v.Next.After(now) {
						s.startJob(v.Job)
						v.Prev = v.Next
						v.Next = v.Schedule.Next(now)
						s.logger.EntryUpdated(v)
						if v.Next.IsZero() {
							s.removeEntry(v.ID)
						}
						s.logger.Info("jobRunning", "id", v.ID, "now", now, "next", v.Next)
					}
				}
			case newEntry := <-s.add:
				timer.Stop()
				now = s.now()
				newEntry.Next = newEntry.Schedule.Next(now)
				if !newEntry.Next.IsZero() {
					s.addEntry(newEntry)
					s.logger.Info("jobAdded", "id", newEntry.ID)
				}
			case id := <-s.remove:
				timer.Stop()
				now = s.now()
				s.removeEntry(id)
				s.logger.Info("jobRemoved", "id", id)
			case c := <-s.snap:
				c <- s.getEntries()
				s.logger.Info("snap")
				continue
			case <-s.stop:
				timer.Stop()
				s.logger.Info("stop")
				return
			}
			break
		}
	}
}

func (s *Scheduler) startJob(job Job) {
	go func() {
		job.Run()
	}()
}
