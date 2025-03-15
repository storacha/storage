// Copyright (c) https://github.com/maragudk/goqite
// https://github.com/maragudk/goqite/blob/6d1bf3c0bcab5a683e0bc7a82a4c76ceac1bbe3f/LICENSE
//
// This source code is licensed under the MIT license found in the LICENSE file
// in the root directory of this source tree, or at:
// https://opensource.org/licenses/MIT

// Package jobs provides a [JobQueue] which can run registered job [Func]s by name, when a message for it is received
// on the underlying queue.
//
// It provides:
//   - Limit on how many jobs can be run simultaneously
//   - Automatic message timeout extension while the job is running
//   - Graceful shutdown
package jobqueue

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/storacha/storage/pkg/pdp/aggregator/jobqueue/queue"
)

type JobQueue[T any] struct {
	queue         *queue.Queue
	jobs          map[string]func(ctx context.Context, msg T) error
	pollInterval  time.Duration
	extend        time.Duration
	jobCount      int
	jobCountLimit int
	jobCountLock  sync.RWMutex
	log           StandardLogger
	serializer    Serializer[T]
}

func New[T any](q *queue.Queue, ser Serializer[T], options ...Option) *JobQueue[T] {
	// Default config
	cfg := &Config{
		Log:           &discardLogger{},
		JobCountLimit: runtime.GOMAXPROCS(0),
		PollInterval:  100 * time.Millisecond,
		Extend:        5 * time.Second,
	}

	// Apply all provided options to the config
	for _, opt := range options {
		opt(cfg)
	}

	// Construct the JobQueue using the final config
	jq := &JobQueue[T]{
		jobs: make(map[string]func(ctx context.Context, msg T) error),

		queue:      q,
		serializer: ser,

		log:           cfg.Log,
		jobCountLimit: cfg.JobCountLimit,
		pollInterval:  cfg.PollInterval,
		extend:        cfg.Extend,
	}
	return jq
}

type message struct {
	Name    string
	Message []byte
}

// Start the JobQueue, blocking until the given context is cancelled.
// When the context is cancelled, waits for the jobs to finish.
func (r *JobQueue[T]) Start(ctx context.Context) {
	var names []string
	for k := range r.jobs {
		names = append(names, k)
	}
	sort.Strings(names)

	r.log.Infow("Starting", "jobs", names)

	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			r.log.Infow("Stopping")
			wg.Wait()
			r.log.Infow("Stopped")
			return
		default:
			r.receiveAndRun(ctx, &wg)
		}
	}
}

func (r *JobQueue[T]) Register(name string, fn func(ctx context.Context, msg T) error) {
	if _, ok := r.jobs[name]; ok {
		panic(fmt.Sprintf(`job "%v" already registered`, name))
	}
	r.jobs[name] = fn
}

func (r *JobQueue[T]) Enqueue(ctx context.Context, name string, msg T) error {
	r.log.Infof("Enqueue -> %s: %v", name, msg)
	m, err := r.serializer.Serialize(msg)
	if err != nil {
		return fmt.Errorf("serializer error: %w", err)
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(message{Name: name, Message: m}); err != nil {
		return err
	}
	return r.queue.Send(ctx, queue.Message{Body: buf.Bytes()})
}

func (r *JobQueue[T]) EnqueueTx(ctx context.Context, tx *sql.Tx, name string, msg T) error {
	m, err := r.serializer.Serialize(msg)
	if err != nil {
		return fmt.Errorf("serializer error: %w", err)
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(message{Name: name, Message: m}); err != nil {
		return err
	}
	return r.queue.SendTx(ctx, tx, queue.Message{Body: buf.Bytes()})
}

func (r *JobQueue[T]) receiveAndRun(ctx context.Context, wg *sync.WaitGroup) {
	r.jobCountLock.RLock()
	if r.jobCount == r.jobCountLimit {
		r.jobCountLock.RUnlock()
		// This is to avoid a busy loop
		time.Sleep(r.pollInterval)
		return
	} else {
		r.jobCountLock.RUnlock()
	}

	m, err := r.queue.ReceiveAndWait(ctx, r.pollInterval)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return
		}
		r.log.Errorw("Error receiving job", "error", err)
		// Sleep a bit to not hammer the queue if there's an error with it
		time.Sleep(time.Second)
		return
	}

	if m == nil {
		return
	}

	var jm message
	if err := json.NewDecoder(bytes.NewReader(m.Body)).Decode(&jm); err != nil {
		r.log.Errorw("Error decoding job message body", "error", err)
		return
	}

	jobInput, err := r.serializer.Deserialize(jm.Message)
	if err != nil {
		r.log.Errorw("Error deserializing job message", "error", err)
		return
	}

	r.log.Infof("Dequeue -> %s: %v", jm.Name, jobInput)
	job, ok := r.jobs[jm.Name]
	if !ok {
		panic(fmt.Sprintf(`job "%v" not registered`, jm.Name))
	}

	r.jobCountLock.Lock()
	r.jobCount++
	r.jobCountLock.Unlock()

	wg.Add(1)
	go func() {
		defer wg.Done()

		defer func() {
			r.jobCountLock.Lock()
			r.jobCount--
			r.jobCountLock.Unlock()
		}()

		defer func() {
			if rec := recover(); rec != nil {
				r.log.Errorw("Recovered from panic in job", "error", rec)
			}
		}()

		jobCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		// Extend the job message while the job is running
		go func() {
			// Start by sleeping so we don't extend immediately
			time.Sleep(r.extend - r.extend/5)
			for {
				select {
				case <-jobCtx.Done():
					return
				default:
					r.log.Infow("Extending message timeout", "name", jm.Name)
					if err := r.queue.Extend(jobCtx, m.ID, r.extend); err != nil {
						r.log.Errorw("Error extending message timeout", "error", err)
					}
					time.Sleep(r.extend - r.extend/5)
				}
			}
		}()

		r.log.Infow("Running job", "name", jm.Name)
		before := time.Now()
		if err := job(jobCtx, jobInput); err != nil {
			r.log.Errorw("Error running job", "name", jm.Name, "error", err)
			return
		}
		duration := time.Since(before)
		r.log.Errorw("Ran job", "name", jm.Name, "duration", duration)

		deleteCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// TODO(we don't want to retry failures here, this should be rare, but worth fixing.
		if err := r.queue.Delete(deleteCtx, m.ID); err != nil {
			r.log.Errorw("Error deleting job from queue, it will be retried", "error", err)
		}
	}()
}
