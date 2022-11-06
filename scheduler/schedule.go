package scheduler

import "time"

type Schedule interface {
	Next(time.Time) time.Time
}

type OneTimeSchedule struct {
	T time.Time
}

func (s *OneTimeSchedule) Next(t time.Time) time.Time {
	r := s.T
	s.T = time.Time{}
	return r
}
