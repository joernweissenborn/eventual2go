package eventual2go

import (
	"errors"
	"time"
)

// ErrTimeout represents a timeout error
var ErrTimeout = errors.New("Timeout")

// Completer is thread-safe struct that can be completed with arbitrary data or failed with an error. Handler functions can
// be registered for both events and get invoked after completion..
type Completer struct {
	f *Future
}

// NewCompleter creates a new Completer.
func NewCompleter() (c *Completer) {
	c = &Completer{newFuture()}
	return
}

// NewTimeoutCompleter creates a new Completer, which error completes after the specified duration, if Completer hasnt been completed otherwise.
func NewTimeoutCompleter(d time.Duration) (c *Completer) {
	c = &Completer{newFuture()}

	go timeout(c, d)

	return
}

// Complete completes the Completer with the given data and triggers all registered completion handlers. Panics if the Completer is already complete.
func (c *Completer) Complete(d Data) {
	c.f.complete(d)
}

// Future returns the associated Future.
func (c *Completer) Future() (f *Future) {
	return c.f
}

// Completed returns the completion state.
func (c *Completer) Completed() bool {
	return c.f.Completed()
}

// CompleteError completes the Completer with the given error and triggers all registered error handlers. Panics if the Completer is already complete.
func (c *Completer) CompleteError(err error) {
	c.f.completeError(err)
}

func (c *Completer) CompleteOnFuture(f *Future) {
	f.Then(completeFuture(c))
	f.Err(completeFutureError(c))
}

func completeFuture(c *Completer) CompletionHandler {
	return func(d Data) Data {
		c.Complete(d)
		return nil
	}
}

func completeFutureError(c *Completer) ErrorHandler {
	return func(err error) (Data, error) {
		c.CompleteError(err)
		return nil, nil
	}
}

func timeout(c *Completer, d time.Duration) {
	time.Sleep(d)
	if !c.Completed() {
		c.CompleteError(ErrTimeout)
	}
}
