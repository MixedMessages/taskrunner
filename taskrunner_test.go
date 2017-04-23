package taskrunner

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"

	"golang.org/x/sync/errgroup"
)

var errMockFailure = errors.New("mock op failure")

type mockTask struct {
	withError bool
	task      func(context.Context) (interface{}, error)
}

func (w *mockTask) Task(ctx context.Context) (interface{}, error) {
	if w.withError {
		return nil, errMockFailure
	}

	return w.task(ctx)
}

func TestNewTaskRunner(t *testing.T) {
	tests := []struct {
		name string

		options []func(*TaskRunner) error
		wantErr bool
	}{
		{
			"Pool Start/Teardown",
			[]func(*TaskRunner) error{},
			false,
		},
		{
			"FAIL - Invalid Amount Of Workers",
			[]func(*TaskRunner) error{OptionMaxGoroutines(-1)},
			true,
		},
		{
			"Success - Valid goroutines",
			[]func(*TaskRunner) error{OptionMaxGoroutines(100)},
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewTaskRunner(test.options...)
			if (err != nil && !test.wantErr) || (err == nil && test.wantErr) {
				t.Errorf("unexpected result creating taskrunner - expected result=%v - err=%v\n", test.wantErr, err)
			}
		})
	}
}

func TestTaskRunnerRun(t *testing.T) {
	timeout := time.Duration(10) * time.Millisecond

	tests := []struct {
		name    string
		workers int
		runs    int
		task    func(context.Context) (interface{}, error)
		wantErr bool
	}{
		{
			"Success - Lots of workers",
			10,
			1,
			func(context.Context) (interface{}, error) {
				time.Sleep(timeout / 2)
				return nil, nil
			},
			false,
		},
		{
			"Failure - Too many tasks, task too slow, exceeded timeout",
			1,
			2,
			func(context.Context) (interface{}, error) {
				time.Sleep(timeout * 2)
				return nil, nil
			},
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner, err := NewTaskRunner(OptionMaxGoroutines(test.workers))
			if err != nil {
				t.Errorf("unexpected error creating taskrunner - err=%v\n", err)
				t.FailNow()
			}

			if err := runner.Start(); err != nil {
				t.Errorf("unexpected error starting taskrunner - err=%+v", err)
				t.FailNow()
			}

			var runError error
			for i := 0; i < test.runs; i++ {

				ctx, done := context.WithTimeout(context.Background(), timeout)
				promise := runner.Run(ctx, &mockTask{false, test.task})

				if _, err := promise(); err != nil {
					runError = err
				}

				done()
			}

			if (runError != nil && !test.wantErr) || (runError == nil && test.wantErr) {
				t.Errorf("unexpected result running task - expected result=%v - err=%v\n", test.wantErr, runError)
			}

			if err := runner.Stop(); err != nil {
				t.Errorf("unexpected error stopping taskrunner - err=%+v", err)
				t.FailNow()
			}
		})
	}
}

func TestTaskRunnerMultipleStartStop(t *testing.T) {

	runner, err := NewTaskRunner()
	if err != nil {
		t.Errorf("unexpected error creating taskrunner - err=%v\n", err)
		t.FailNow()
	}

	var eg errgroup.Group

	for i := 0; i < 10; i++ {
		eg.Go(runner.Start)
	}

	if err := eg.Wait(); err == nil {
		t.Error("unexpected nil error from multiple attempted starts")
	}

	for i := 0; i < 10; i++ {
		eg.Go(runner.Stop)
	}

	if err := eg.Wait(); err == nil {
		t.Error("unexpected nil error from multiple attempted stops")
	}

	for i := 0; i < 10; i++ {
		eg.Go(runner.Start)
		eg.Go(runner.Stop)
	}

	if err := eg.Wait(); err == nil {
		t.Errorf("unexpected nil error from concurrency start-stop of runner")
	}
}

func TestTaskRunnerConcurrentStartRunStop(t *testing.T) {

	runner, err := NewTaskRunner()
	if err != nil {
		t.Errorf("unexpected error creating taskrunner - err=%v\n", err)
		t.FailNow()
	}

	if err := runner.Start(); err != nil {
		t.Errorf("unexpected error starting taskrunner - err=%+v", err)
		t.FailNow()
	}

	var eg errgroup.Group

	for i := 0; i < 100; i++ {
		eg.Go(func() error {
			promise := runner.Run(context.TODO(), &mockTask{false, func(context.Context) (interface{}, error) {
				time.Sleep(time.Duration(10) * time.Millisecond)
				return "test", nil
			}})

			_, err := promise()

			return err
		})
	}

	if err := eg.Wait(); err != nil {
		t.Errorf("unexpected error from concurrency start-run-stop of runner - err=%+v", err)
	}

	if err := runner.Stop(); err != nil {
		t.Errorf("unexpected error from stopping task runner - err=%+v", err)
	}

}
