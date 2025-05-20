package scheduler

import (
	"time"

	"gorm.io/gorm"
)

// TaskID represents the task identifier.
type TaskID int64

// AddTaskFunc is used to add extra information when creating a task.
type AddTaskFunc func(extraInfo func(TaskID, *gorm.DB) (shouldCommit bool, seriousError error))

// TaskInterface defines what a task must implement.
type TaskInterface interface {
	Do(taskID TaskID) (done bool, err error)
	TypeDetails() TaskTypeDetails
	Adder(AddTaskFunc)
}

// TaskTypeDetails defines static properties for each task type.
type TaskTypeDetails struct {
	// Task name (should be unique and short)
	Name string
	// Maximum failure count before dropping the task (0 = retry forever)
	MaxFailures uint
	// RetryWait is a function returning the wait duration based on retries.
	RetryWait func(retries int) time.Duration
	// PeriodicScheduler defines a task that should run on a fixed interval
	PeriodicScheduler *PeriodicScheduler
}

// PeriodicScheduler defines a periodic task scheduler that runs on a fixed interval
type PeriodicScheduler struct {
	// Interval is the time between executions
	Interval time.Duration
	// Runner is the function that will be called periodically
	Runner func(AddTaskFunc) error
}
