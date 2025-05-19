package scheduler

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/jackc/pgconn"
	"gorm.io/gorm"

	"github.com/storacha/storage/pkg/pdp/service/models"
)

// taskTypeHandler ties a task implementation with engine-specific metadata.
type taskTypeHandler struct {
	TaskInterface
	TaskTypeDetails TaskTypeDetails
	TaskEngine      *TaskEngine
	// Additional fields like concurrency limiters can be added here.
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
			AddedBy:    h.TaskEngine.sessionID,
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

		return ErrDoNotCommit
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

		if errors.Is(err, ErrDoNotCommit) {
			return
		}
		log.Errorw("Could not add task. AddTask func failed", "error", err, "type", h.TaskTypeDetails.Name)
		return
	}

}

// considerWork claims and executes tasks.
func (h *taskTypeHandler) considerWork(taskIDs []TaskID, db *gorm.DB) bool {
	acceptedAny := false

	for _, id := range taskIDs {
		result := db.Model(&models.Task{}).
			Where(&models.Task{ID: int64(id), SessionID: nil}).
			Updates(models.Task{
				SessionID:  &h.TaskEngine.sessionID,
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
		// TODO doing this in parallel is causing concurrency issues with the sqlite database.
		acceptedAny = true
		func(taskID TaskID) {
			var (
				done    bool
				doErr   error
				doStart = time.Now()
			)
			defer func() {
				if r := recover(); r != nil {
					stackSlice := make([]byte, 4092)
					sz := runtime.Stack(stackSlice, false)
					log.Error("Recovered from a serious error "+
						"while processing "+h.TaskTypeDetails.Name+" task "+strconv.Itoa(int(taskID))+": ", r,
						" Stack: ", string(stackSlice[:sz]))
				}

				h.handleDoneTask(taskID, doStart, done, doErr)
			}()

			log.Infow(
				"Executing task",
				"name", h.TaskTypeDetails.Name,
				"id", taskID,
				"session", h.TaskEngine.sessionID,
			)
			done, doErr = h.Do(taskID)
			if doErr != nil {
				log.Errorw("Error executing task", "task_id", taskID, "error", doErr)
			}
		}(id)
	}

	return acceptedAny
}

func (h *taskTypeHandler) handleDoneTask(id TaskID, startTime time.Time, done bool, doErr error) error {
	endTime := time.Now()
	err := h.TaskEngine.db.WithContext(h.TaskEngine.ctx).Debug().Transaction(func(tx *gorm.DB) error {

		// find the task that we are handling
		task := models.Task{}
		if res := tx.Model(&models.Task{}).
			Where("id = ?", id).
			First(&task); res.Error != nil {
			return fmt.Errorf("failed to handle done task: failed to query taskID: %d: %w", id, res.Error)
		} else if res.RowsAffected == 0 {
			return fmt.Errorf("failed to handle done task: no task found for taskID: %d: %w", id, res.Error)
		}

		taskErrMsg := ""
		if done {
			// if the task is done, we can delete it
			if err := tx.Model(&models.Task{}).Delete(task).Error; err != nil {
				return fmt.Errorf("failed to handle done task: failed to delete task %d: %w", id, err)
			}
			// the task may have returned an error, in addition to completing successfully, record this if present
			if doErr != nil {
				taskErrMsg = "non-failing error: " + doErr.Error()
			}
		} else {
			// if the task is not done, see if it can be retried, and capture its error message
			if doErr != nil {
				taskErrMsg = "error: " + doErr.Error()
			}
			// the task has exceeded the number of allowed retries, delete it
			if h.TaskTypeDetails.MaxFailures > 0 && task.Retries >= h.TaskTypeDetails.MaxFailures {
				if err := tx.Delete(&models.Task{}, id).Error; err != nil {
					return fmt.Errorf("failed to deleted failed task %d: %w", id, err)
				}
			} else {
				// the task may be retried, increment retry counter and set sessionID to nil, allowing the task engine
				// to pick it back up and try again
				if err := tx.Model(&models.Task{}).
					Where(&models.Task{ID: int64(id)}).
					Updates(models.Task{
						SessionID:  nil,
						Retries:    task.Retries + 1,
						UpdateTime: time.Now(),
					}).Error; err != nil {
					return fmt.Errorf("failed to updated failed task %d: %w", id, err)
				}
			}
		}

		// record history about the task.
		if res := tx.Create(&models.TaskHistory{
			TaskID:               int64(id),
			Name:                 task.Name,
			Posted:               task.PostedTime,
			WorkStart:            startTime,
			WorkEnd:              endTime,
			Result:               done,
			Err:                  taskErrMsg,
			CompletedBySessionID: h.TaskEngine.sessionID,
		}); res.Error != nil {
			return fmt.Errorf("failed to write task (%d) history: %w", id, res.Error)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

var ErrDoNotCommit = errors.New("do not commit")

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
