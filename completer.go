package eventual2go

import (
	"errors"
	"time"
)

// Completer is thread-safe struct that can be completed with arbitrary data or failed with an error. Handler functions can
// be registered for both events and get invoked after completion..
type Completer struct {
	f *Future
}

// Creates a new Completer.
func NewCompleter() (c *Completer) {
	c = &Completer{NewFuture()}
	return
}

// Creates a new Completer, which error completes after the specified duration, if Completer hasnt been completed otherwise.
func NewTimeoutCompleter(d time.Duration) (c *Completer) {
	c = &Completer{NewFuture()}

	go timeout(c, d)

	return
}

// Completes the Completer with the given data and triggers all registered completion handlers. Panics if the Completer is already
// complete.
func (c *Completer) Complete(d Data) {
	c.f.complete(d)
}

// Returns the associated future.
func (c *Completer) Future() (f *Future) {
	return c.f
}

// Returns the completion state.
func (c *Completer) Completed() bool {
	return c.f.completed
}

// Completes the Completer with the given error and triggers all registered error handlers. Panics if the Completer is already
// complete.
func (c *Completer) CompleteError(err error) {
	c.f.completeError(err)
}

func timeout(c *Completer, d time.Duration) {
	time.Sleep(d)
	if !c.Completed() {
		c.CompleteError(errors.New("Timeout"))
	}
}
