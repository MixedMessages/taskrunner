package taskrunner

import (
	"errors"

	"github.com/go-kit/kit/metrics"
)

// OptionMaxGoroutines is a functional option for configuring the number of
// workers in a TaskRunner.
func OptionMaxGoroutines(n int) Option {
	return func(r *TaskRunner) error {
		if n <= 0 {
			return errors.New("number of goroutines must be postive")
		}

		r.maxWorkers = n
		return nil
	}
}

// OptionTaskCounter allows access to the a metrics.Counter which aggregates
// the number of tasks processed.
func OptionTaskCounter(ctr metrics.Counter) Option {
	return func(r *TaskRunner) error {
		if ctr == nil {
			return errors.New("counter must be non-nil")
		}

		r.taskCounter = ctr
		return nil
	}
}

// OptionUnhandledPromisesGauge allows a go-kit metrics.Gauge to be passed-in
// collect the number of unhandled promises.
// Useful to discover if there is a leak of unhandled promises in-memory.
func OptionUnhandledPromisesGauge(gauge metrics.Gauge) Option {
	return func(r *TaskRunner) error {
		if gauge == nil {
			return errors.New("gauge must be non-nil")
		}

		r.unhandledPromisesGauge = gauge
		return nil
	}
}

// OptionWorkersGauge allows access to the current number of workers via a
// go-kit metrics.Gauge.
func OptionWorkersGauge(gauge metrics.Gauge) Option {
	return func(r *TaskRunner) error {
		if gauge == nil {
			return errors.New("gauge must be non-nil")
		}

		r.workersGauge = gauge
		return nil
	}
}

// OptionTaskTimeHistogram allows access to customize a histogram for sampling
// average task times.
func OptionTaskTimeHistogram(histogram metrics.Histogram) Option {
	return func(r *TaskRunner) error {
		if histogram == nil {
			return errors.New("histogram must be non-nil")
		}

		r.averageTaskTime = histogram
		return nil
	}
}
