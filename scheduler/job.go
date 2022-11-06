package scheduler

import (
	"sync"
	"time"
)

type Job interface {
	Run()
}

type FuncJob func()

func (f FuncJob) Run() {
	f()
}

type Warpper func(Job) Job

func DelayIfStillRunnig(logger Logger) Warpper {
	return func(j Job) Job {
		var mu sync.Mutex
		return FuncJob(func() {
			s := time.Now()
			mu.Lock()
			defer mu.Unlock()
			if dur := time.Since(s); dur > time.Second*10 {
				logger.Info("delay", "duration", dur)
			}
			j.Run()
		})
	}
}

func SkipIfStillRunning(logger Logger) Warpper {
	return func(j Job) Job {
		ch := make(chan struct{}, 1)
		ch <- struct{}{}
		return FuncJob(func() {
			select {
			case <-ch:
				defer func() { ch <- struct{}{} }()
				j.Run()
			default:
				logger.Info("skip")
			}
		})
	}
}
