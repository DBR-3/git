package workers

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/stock-market-system/pkg/logger"
)

// Task represents a work task
type Task struct {
	ID      string
	Ticker  string
	Process func(ctx context.Context) error
}

// Result represents task result
type Result struct {
	TaskID  string
	Ticker  string
	Success bool
	Error   error
	Data    interface{}
}

// WorkerPool manages a pool of workers
type WorkerPool struct {
	numWorkers     int
	taskQueue      chan Task
	resultQueue    chan Result
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	logger         *logger.Logger
	tasksProcessed uint64
	tasksFailed    uint64
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(numWorkers int, taskQueueSize int, logger *logger.Logger) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		numWorkers:  numWorkers,
		taskQueue:   make(chan Task, taskQueueSize),
		resultQueue: make(chan Result, taskQueueSize),
		ctx:         ctx,
		cancel:      cancel,
		logger:      logger,
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	wp.logger.Infof("Starting worker pool with %d workers", wp.numWorkers)

	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// worker processes tasks from the queue
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	wp.logger.Debugf("Worker %d started", id)

	for {
		select {
		case <-wp.ctx.Done():
			wp.logger.Debugf("Worker %d stopping", id)
			return

		case task, ok := <-wp.taskQueue:
			if !ok {
				wp.logger.Debugf("Worker %d: task queue closed", id)
				return
			}

			wp.logger.Debugf("Worker %d processing task %s (ticker: %s)", id, task.ID, task.Ticker)

			startTime := time.Now()
			err := task.Process(wp.ctx)
			duration := time.Since(startTime)

			result := Result{
				TaskID:  task.ID,
				Ticker:  task.Ticker,
				Success: err == nil,
				Error:   err,
			}

			if err != nil {
				atomic.AddUint64(&wp.tasksFailed, 1)
				wp.logger.Errorf("Worker %d: task %s failed after %v: %v", id, task.ID, duration, err)
			} else {
				atomic.AddUint64(&wp.tasksProcessed, 1)
				wp.logger.Debugf("Worker %d: task %s completed in %v", id, task.ID, duration)
			}

			// Send result
			select {
			case wp.resultQueue <- result:
			case <-wp.ctx.Done():
				return
			}
		}
	}
}

// Submit submits a task to the worker pool
func (wp *WorkerPool) Submit(task Task) error {
	select {
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is stopped")
	case wp.taskQueue <- task:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout submitting task")
	}
}

// Results returns the result channel
func (wp *WorkerPool) Results() <-chan Result {
	return wp.resultQueue
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() {
	wp.logger.Info("Stopping worker pool...")

	// Close task queue
	close(wp.taskQueue)

	// Wait for all workers to finish
	wp.wg.Wait()

	// Cancel context
	wp.cancel()

	// Close result queue
	close(wp.resultQueue)

	wp.logger.Infof("Worker pool stopped. Processed: %d, Failed: %d",
		wp.tasksProcessed, wp.tasksFailed)
}

// GetStats returns worker pool statistics
func (wp *WorkerPool) GetStats() (processed, failed uint64) {
	return atomic.LoadUint64(&wp.tasksProcessed), atomic.LoadUint64(&wp.tasksFailed)
}

// Wait waits for all tasks to complete with timeout
func (wp *WorkerPool) Wait(timeout time.Duration) error {
	done := make(chan struct{})

	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout waiting for tasks to complete")
	}
}
