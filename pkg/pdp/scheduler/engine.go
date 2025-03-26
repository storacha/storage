package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/jackc/pgconn"
	"gorm.io/gorm"

	"github.com/storacha/storage/pkg/pdp/service/models"
)

var log = logging.Logger("pdp/scheduler")

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

// Every is a helper function that will call the provided callback
// function at most once every `passEvery` duration. If the function is called
// more frequently than that, it will return nil and not call the callback.
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

// TaskInterface defines what a task must implement.
type TaskInterface interface {
	Do(taskID TaskID) (done bool, err error)
	TypeDetails() TaskTypeDetails
	Adder(AddTaskFunc)
}

// AddTaskFunc is used to add extra information when creating a task.
type AddTaskFunc func(extraInfo func(TaskID, *gorm.DB) (shouldCommit bool, seriousError error))

// TaskID represents the task identifier.
type TaskID int

// TaskEngine is the central scheduler.
type TaskEngine struct {
	ctx            context.Context
	cancel         context.CancelFunc
	db             *gorm.DB
	owner          int
	handlers       []*taskTypeHandler
	taskMap        map[string]*taskTypeHandler
	lastFollowTime time.Time
	lastCleanup    time.Time
}

// taskTypeHandler ties a task implementation with engine-specific metadata.
type taskTypeHandler struct {
	TaskInterface
	TaskTypeDetails TaskTypeDetails
	TaskEngine      *TaskEngine
	// Additional fields like concurrency limiters can be added here.
}

// New creates a new TaskEngine with the provided task implementations.
func NewEngine(db *gorm.DB, impls []TaskInterface) (*TaskEngine, error) {
	ctx, cancel := context.WithCancel(context.Background())
	e := &TaskEngine{
		ctx:     ctx,
		owner:   1,
		cancel:  cancel,
		db:      db,
		taskMap: make(map[string]*taskTypeHandler, len(impls)),
	}

	for _, impl := range impls {
		h := &taskTypeHandler{
			TaskInterface:   impl,
			TaskTypeDetails: impl.TypeDetails(),
			TaskEngine:      e,
		}
		e.handlers = append(e.handlers, h)
		e.taskMap[h.TaskTypeDetails.Name] = h

		// Start the adder routine for the task type.
		go h.Adder(h.AddTask)

		// **Start the periodic "bored" routine if provided**
		if h.TaskTypeDetails.IAmBored != nil {
			go func(h *taskTypeHandler) {
				err := h.TaskTypeDetails.IAmBored(h.AddTask)
				if err != nil {
					log.Warnf("IAmBored for task %s returned error: %v",
						h.TaskTypeDetails.Name, err)
				}
			}(h)
		}
	}

	go e.poller()

	return e, nil
}

// GracefullyTerminate stops new task scheduling.
func (e *TaskEngine) GracefullyTerminate() {
	e.cancel()
	// Optionally wait for currently running tasks to finish.
}

// poller continuously checks for work.
func (e *TaskEngine) poller() {
	pollDuration := 3 * time.Second
	pollNextDuration := 100 * time.Millisecond
	nextWait := pollNextDuration

	for {
		select {
		case <-time.After(nextWait):
		case <-e.ctx.Done():
			return
		}
		nextWait = pollDuration

		accepted := e.pollerTryAllWork()
		if accepted {
			nextWait = pollNextDuration
		}
		// Here you could also call a follow-up work routine if needed.
	}
}

const (
	WorkSourcePoller   = "poller"
	WorkSourceRecover  = "recovered"
	WorkSourceIAmBored = "bored"
)

// pollerTryAllWork looks for unassigned tasks in the DB and schedules them.
func (e *TaskEngine) pollerTryAllWork() bool {
	// Optional cleanup logic.
	if time.Since(e.lastCleanup) > 5*time.Minute {
		e.lastCleanup = time.Now()
		// (Cleanup code can be added here if needed.)
	}

	// Iterate over all registered task types.
	for _, h := range e.handlers {
		var tasks []models.Task
		// Fetch tasks for this type that are unassigned.
		// (Assuming unassigned tasks are filtered by name; add additional conditions if needed.)
		if err := e.db.WithContext(e.ctx).
			Where("name = ? AND owner_id IS NULL", h.TaskTypeDetails.Name).
			Order("update_time").
			Find(&tasks).Error; err != nil {
			log.Errorf("Unable to read work for task type %s: %v", h.TaskTypeDetails.Name, err)
			continue
		}

		var taskIDs []TaskID
		// Filter tasks based on retry logic.
		// Since the Task model no longer has a retries field, we assume a default value (e.g., 0)
		// or adjust the logic if you retrieve retries from another source.
		for _, t := range tasks {
			if h.TaskTypeDetails.RetryWait == nil ||
				time.Since(t.UpdateTime) > h.TaskTypeDetails.RetryWait(0) {
				taskIDs = append(taskIDs, TaskID(t.ID))
			}
		}

		if len(taskIDs) > 0 {
			accepted := h.considerWork(WorkSourcePoller, taskIDs, e.db)
			if accepted {
				return true
			}
			log.Warnf("Work not accepted for %d %s task(s)", len(taskIDs), h.TaskTypeDetails.Name)
		}
	}

	// if no work was accepted, are we bored? Then find work in priority order.
	for _, v := range e.handlers {
		v := v
		if v.TaskTypeDetails.IAmBored != nil {
			var added []TaskID
			err := v.TaskTypeDetails.IAmBored(func(extraInfo func(TaskID, *gorm.DB) (shouldCommit bool, seriousError error)) {
				v.AddTask(func(tID TaskID, tx *gorm.DB) (shouldCommit bool, seriousError error) {
					b, err := extraInfo(tID, tx)
					if err == nil && shouldCommit {
						added = append(added, tID)
					}
					return b, err
				})
			})
			if err != nil {
				log.Error("IAmBored failed: ", err)
				continue
			}
			if added != nil { // tiny chance a fail could make these bogus, but considerWork should then fail.
				v.considerWork(WorkSourceIAmBored, added, e.db)
			}
		}
	}

	return false
}

// AddTask is the implementation passed to each task's Adder.
// It creates a new task record in the database.
func (h *taskTypeHandler) AddTask(extra func(TaskID, *gorm.DB) (bool, error)) {
	var tID TaskID
	retryWait := 100 * time.Millisecond

retryAddTask:
	err := h.TaskEngine.db.WithContext(h.TaskEngine.ctx).Transaction(func(tx *gorm.DB) error {
		// Create a new Task record.
		task := models.Task{
			PostedTime: time.Now(),
			AddedBy:    h.TaskEngine.owner, // For a single worker, this might be 0 or some constant.
			Name:       h.TaskTypeDetails.Name,
		}

		// Insert the task and let GORM fill in the auto-generated ID.
		if err := tx.Create(&task).Error; err != nil {
			return fmt.Errorf("could not insert task: %w", err)
		}

		// Set the task ID from the newly inserted record.
		tID = TaskID(task.ID)

		// Call the extra callback to update additional info in the same transaction.
		shouldCommit, err := extra(tID, tx)
		if err != nil {
			return err
		}

		if shouldCommit {
			return nil
		}

		return DoNotCommitErr
	})
	if err != nil {
		// If a unique constraint error is detected, assume the task already exists.
		if IsUniqueConstraintError(err) {
			log.Debugf("addtask(%s) saw unique constraint, so it's added already.", h.TaskTypeDetails.Name)
			return
		}
		// If it's a serialization error, backoff and retry.
		if IsSerializationError(err) {
			time.Sleep(retryWait)
			retryWait *= 2
			goto retryAddTask
		}

		if errors.Is(err, DoNotCommitErr) {
			return
		}
		log.Errorw("Could not add task. AddTask func failed", "error", err, "type", h.TaskTypeDetails.Name)
		return
	}

}

// considerWork claims and executes tasks.
// In this simplified version, it directly calls the task's Do() method.
func (h *taskTypeHandler) considerWork(source string, taskIDs []TaskID, db *gorm.DB) bool {
	acceptedAny := false

	for _, id := range taskIDs {
		// Attempt to claim ownership of this task in the DB.
		// If RowsAffected == 0, it means another thread or process already took it.
		result := db.Model(&models.Task{}).
			Where("id = ? AND owner_id IS NULL", id).
			Updates(models.Task{
				OwnerID:    &h.TaskEngine.owner, // or a constant, e.g. 1
				UpdateTime: time.Now(),
			})

		if result.Error != nil {
			log.Errorw("Could not claim task", "task_id", id, "error", result.Error)
			continue
		}
		if result.RowsAffected == 0 {
			// Already taken by someone else (or in race condition). Skip it.
			log.Debugf("Task %d was already claimed; skipping", id)
			continue
		}

		// Successfully claimed this task, so let’s run it in a goroutine:
		acceptedAny = true
		go func(taskID TaskID) {
			log.Infow("Executing task", "name", h.TaskTypeDetails.Name, "id", taskID)

			done, err := h.Do(taskID)
			if err != nil {
				log.Errorw("Error executing task", "task_id", taskID, "error", err)
			}

			// If done, remove from DB. Otherwise, release if you want to let it retry:
			if done {
				if err := db.Delete(&models.Task{}, taskID).Error; err != nil {
					log.Errorw("Could not delete completed task", "task_id", taskID, "error", err)
				} else {
					log.Infow("Task completed", "task_id", taskID, "name", h.TaskTypeDetails.Name)
				}
			} else {
				// TODO: in the event of a node failure we need to "un-own" tasks else
				// active tasks during the failure will never be re-run.
				if err := db.Model(&models.Task{}).
					Where("id = ?", taskID).
					Updates(models.Task{OwnerID: nil, UpdateTime: time.Now()}).Error; err != nil {
					log.Errorw("Could not release task", "task_id", taskID, "error", err)
				}
			}
		}(id)
	}

	return acceptedAny
}

var DoNotCommitErr = errors.New("do not commit")

func IsUniqueConstraintError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// 23505 is the PostgreSQL error code for unique violation.
		return pgErr.Code == "23505"
	}
	return false
}

func IsSerializationError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// 40001 is the PostgreSQL error code for serialization failure.
		return pgErr.Code == "40001"
	}
	return false
}
