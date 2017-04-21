// Package taskrunner provides an API for running concurrent tasks in a promise-like
// style without having to deal with concurrency directly.
//
// The library provides an interface that must be implemented to run a Task concurrently
// and a concurrency-safe API for running tasks concurrently without managing the
// channels, goroutines and waitgroups yourself.
package taskrunner
