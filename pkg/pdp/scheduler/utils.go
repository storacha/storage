package scheduler

import (
	"sync"
	"time"
)

// Every is a helper function that will call the provided callback
// function at most once every `passEvery` duration. If the function is called
// more frequently than that, it will return nil and not call the callback.
// Deprecated: Use NewPeriodicScheduler instead.
func Every[P, R any](passInterval time.Duration, cb func(P) R) func(P) R {
	var lastCall time.Time
	var lk sync.Mutex

	return func(param P) R {
		lk.Lock()
		defer lk.Unlock()

		if time.Since(lastCall) < passInterval {
			return *new(R)
		}

		defer func() {
			lastCall = time.Now()
		}()
		return cb(param)
	}
}

// NewPeriodicScheduler creates a new PeriodicScheduler with the given interval and runner function.
// This is a helper function to make it easier to create a PeriodicScheduler for tasks
// that need to run periodically.
func NewPeriodicScheduler(interval time.Duration, runner func(AddTaskFunc) error) *PeriodicScheduler {
	return &PeriodicScheduler{
		Interval: interval,
		Runner:   runner,
	}
}
