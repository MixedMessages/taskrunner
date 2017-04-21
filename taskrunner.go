package taskrunner

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

// Defaults for the TaskRunner.
const (
	maxWorkers = 4
)

// State enum that reflects the internal state of the TaskRunner.
// Used to determine if the runner is running or stopped.
const (
	stoppedState uint32 = iota
	startedState
)

var (

	// errTimeoutExceeded is returned whenever a request is cancelled.
	errTimeoutExceeded = errors.New("timeout exceeded - worker not available to process job")

	// errRunnerNotStarted is given whenever the Runner is in the stopped state
	// and a Run is attempted.
	errRunnerNotStarted = errors.New("runner is not currently started, cannot run tasks")

	// errRunnerAlreadyStarted is given whenever the Runner is in a started state
	// and receives another Start signal.
	errRunnerAlreadyStarted = errors.New("runner is already started and cannot be started again")

	// errRunnerAlreadyStopped signals that the runner is already stooped.
	errRunnerAlreadyStopped = errors.New("runner is already stopped and cannot be stopped again")
)

// Task is an interface for performing a given task.
// Task self-describes how the job is to be handled by the Task.
// Returns an error to report task completion.
type Task interface {
	Task(context.Context) (interface{}, error)
}

// TaskRunner is a Runner capable of concurrently running Tasks.
// Runs multiple goroutines to process Tasks concurrently.
type TaskRunner struct {
	tasks chan taskWrapper
	exit  chan struct{}

	wg sync.WaitGroup

	mtx sync.RWMutex

	state uint32

	maxWorkers int
}

// NewTaskRunner creates an TaskRunner.
// Provides functional options for configuring the TaskRunner while also
// validating the input configurations.
// Returns an error if the TaskRunner is improperly configured.
func NewTaskRunner(options ...func(*TaskRunner) error) (*TaskRunner, error) {

	p := TaskRunner{
		tasks: make(chan taskWrapper),
		exit:  make(chan struct{}),

		state: stoppedState,

		maxWorkers: maxWorkers,
	}

	for _, opt := range options {
		if err := opt(&p); err != nil {
			return nil, err
		}

	}

	return &p, nil
}

// taskWrapper wraps a task with its context and result channel. Provides a single
// payload object to send through channels.
type taskWrapper struct {
	ctx           context.Context
	task          Task
	resultChannel chan taskResult
}

// taskResult is a wrapper struct for the task result. Provides a single payload
// object to send through channels.
type taskResult struct {
	res interface{}
	err error
}

// Run gives the Task to an available worker.
// The given context is used as a hook to cancel a running worker task.
// Run returns a closure over the result of a Task. When the result of a Task
// is desired, you can call the function to retrieve the result.
func (p *TaskRunner) Run(ctx context.Context, w Task) func() (interface{}, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	if !p.isRunning() {
		return func() (interface{}, error) {
			return nil, errRunnerNotStarted
		}
	}

	resultChannel := make(chan taskResult, 1)

	task := taskWrapper{
		ctx:           ctx,
		task:          w,
		resultChannel: resultChannel,
	}

	select {
	case p.tasks <- task:
		// Return a closure over the result channel response.
		return func() (interface{}, error) {

			select {
			case result := <-resultChannel:
				return result.res, result.err
			case <-ctx.Done():

			}

			return nil, errTimeoutExceeded
		}

	case <-ctx.Done():
		// The context deadline time has been exceeded before a worker could
		// pick up the task. Return a promise with an error.
		return func() (interface{}, error) {
			return nil, errTimeoutExceeded
		}
	}
}

// Start prepares the TaskRunner for processing tasks.
// Once started, a TaskRunner is ready to accept Tasks.
func (p *TaskRunner) Start() error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if p.isRunning() {
		return errRunnerAlreadyStarted
	}

	p.state = startedState
	p.tasks = make(chan taskWrapper)
	p.exit = make(chan struct{})

	p.wg.Add(p.maxWorkers)

	// Start the Task workers.
	for i := 0; i < p.maxWorkers; i++ {
		go func() {

			defer p.wg.Done()

			for {
				select {

				case <-p.exit:
					return

				case w := <-p.tasks:
					res, err := w.task.Task(w.ctx)

					select {
					case w.resultChannel <- taskResult{res, err}:
					case <-w.ctx.Done():
					}
				}
			}
		}()
	}

	return nil
}

// Stop performs a graceful shutdown of the Runner.
// In-progress Tasks are given some time to finish before exitting.
func (p *TaskRunner) Stop() error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if !p.isRunning() {
		return errRunnerNotStarted
	}

	p.state = stoppedState

	// Close the exit channel which signals to workers to cleanup.
	close(p.exit)

	// Wait for the workers to all return before proceeding.
	p.wg.Wait()

	// Close the tasks channel since no more workers can be sending on the channel.
	close(p.tasks)

	return nil
}

// isRunning checks the current state of the TaskRunner.
func (p *TaskRunner) isRunning() bool {
	return atomic.LoadUint32(&p.state) == startedState
}
