# Taskrunner [![GoDoc](https://godoc.org/github.com/MixedMessages/taskrunner?status.svg)](https://godoc.org/github.com/MixedMessages/taskrunner) [![Build Status](https://travis-ci.org/MixedMessages/taskrunner.svg?branch=master)](https://travis-ci.org/MixedMessages/taskrunner)

## Background
Package taskrunner provides an API for running concurrent tasks in a promise-like
style without having to deal with concurrency directly.

The library provides an interface that must be implemented to run a Task concurrently
and a concurrency-safe API for running tasks concurrently without managing the
channels, goroutines and waitgroups yourself.

## Using TaskRunner
To use Taskrunner, you'll need some setup.
Define a struct that implements the `Task` interface to use as the payload to Run.

One example of a naive Task could be something like sending emails.

```go
type EmailPayload struct {
    Email string
    Message string
}

func (p *EmailPayload) Task(ctx context.Context) (interface{}, error) {
    if err := p.sendEmail(); err != nil {
        return nil, err
    }

    return nil, nil
}

func (p *EmailPayload) sendEmail() error {
    // send email
    // ...
    // check and return errors
}
```

Once you've implemented the Task, running the Task is as simple as creating a
new Taskrunner, starting it and running your Task.

Run returns a function closure over the result of your Task so that you can
retrieve it at a later time.
Run accepts a context and passes it to your task so that you can control the
deadline of your Task and even pass request-scoped items via the context.

```go
// Configure the number of workers using a functional option.
runner, err := taskrunner.NewTaskRunner(taskrunner.OptionMaxGoroutines(runtime.NUMCPU + 1))
if err != nil {
    panic(err.Error())
}

// Start the runner.
if err := runner.Start(); err != nil {
    panic(err.Error())
}

// Get your promise.
promise := runner.Run(context.TODO(), &EmailPayload{
    Email: "...",
    Message: "...",
}})

// Check the result of your promise.
if _, err := promise(); err != nil {
    log.Errorf("sending email failed - err=%+v", err)
}

```

## Examples
See the [examples](https://github.com/MixedMessages/taskrunner/tree/master/examples) directory
for some codified example usages of `taskrunner`.

* `helloworld` shows how to run simple Tasks with no return value.
* `helloworld` shows how to run simple Tasks with a return value and possible
error state. Shows how to use the promises and how to perform a type assertion
to safely check the return value.
* `expvar` includes an example of accessing the metric hooks available via
the Option functions and uses `expvar` to demonstrate.
