package taskrunner

import "errors"

// OptionMaxGoroutines is a functional option for configuring the number of
// workers in a TaskRunner.
func OptionMaxGoroutines(n int) func(*TaskRunner) error {
	return func(r *TaskRunner) error {
		if n <= 0 {
			return errors.New("number of goroutines must be postive")
		}

		r.maxWorkers = n
		return nil
	}
}
