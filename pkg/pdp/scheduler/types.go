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
	// Maximum concurrent tasks allowed (0 means no limit)
	Max int
	// Task name (should be unique and short)
	Name string
	// Maximum failure count before dropping the task (0 = retry forever)
	MaxFailures uint
	// RetryWait is a function returning the wait duration based on retries.
	RetryWait func(retries int) time.Duration

	IAmBored func(AddTaskFunc) error
}
