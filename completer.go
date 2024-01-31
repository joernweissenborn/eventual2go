package eventual2go

import (
	"errors"
	"time"
)

// ErrTimeout represents a timeout error
var ErrTimeout = errors.New("Timeout")

// Completer is thread-safe struct that can be completed with arbitrary data or failed with an error. Handler functions can
// be registered for both events and get invoked after completion..
type Completer[T any] struct {
	f *Future[T]
}

// NewCompleter creates a new Completer.
func NewCompleter[T any]() (c *Completer[T]) {
	c = &Completer[T]{newFuture[T]()}
	return
}

// NewTimeoutCompleter creates a new Completer, which error completes after the specified duration, if Completer hasnt been completed otherwise.
func NewTimeoutCompleter[T any](d time.Duration) (c *Completer[T]) {
	c = &Completer[T]{newFuture[T]()}

	go timeout(c, d)

	return
}

// Complete completes the Completer with the given data and triggers all registered completion handlers. Panics if the Completer is already complete.
func (c *Completer[T]) Complete(d T) {
	c.f.complete(d)
}

// CompleteOn invokes a CompletionFunc in a go-routine and either completes with the resut or the error if it is not nil. Don't invoke this function more then once to avoid multiple complition panics.
func (c *Completer[T]) CompleteOn(f CompletionFunc[T]) {
	handler := func(f CompletionFunc[T]) {
		d, err := f()
		if err == nil {
			c.Complete(d)
		} else {
			c.CompleteError(err)
		}
	}
	go handler(f)
}

// Future returns the associated Future.
func (c *Completer[T]) Future() (f *Future[T]) {
	return c.f
}

// Completed returns the completion state.
func (c *Completer[T]) Completed() bool {
	return c.f.Completed()
}

// CompleteError completes the Completer with the given error and triggers all registered error handlers. Panics if the Completer is already complete.
func (c *Completer[T]) CompleteError(err error) {
	c.f.completeError(err)
}

// CompleteOnFuture completes the completer with the result or the error of a `Future`.
func (c *Completer[T]) CompleteOnFuture(f *Future[T]) {
	f.Then(completeFuture(c))
	f.Err(completeFutureError(c))
}

func completeFuture[T any](c *Completer[T]) CompletionHandler[T] {
	return func(d T) {
		c.Complete(d)
	}
}

func completeFutureError[T any](c *Completer[T]) ErrorHandler {
	return func(err error)  {
		c.CompleteError(err)
	}
}

func timeout[T any](c *Completer[T], d time.Duration) {
	time.Sleep(d)
	if !c.Completed() {
		c.CompleteError(ErrTimeout)
	}
}
