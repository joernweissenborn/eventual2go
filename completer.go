package eventual2go


// Completer is thread-safe struct that can be completed with arbitrary data or failed with an error. Handler functions can
// be registered for both events and get invoked after completion..
type Completer struct {

	f *Future

}

// Creates a new Completer.
func NewCompleter() (c *Completer) {
	c = new(Completer)
	c.f = NewFuture()
	return
}

// Completes the Completer with the given data and triggers al registered completion handlers. Panics if the Completer is already
// complete.
func (c *Completer) Complete(d Data){
	c.f.complete(d)
}

// Returns the associated future.
func (c *Completer) Future()(f *Future){
	return c.f
}

// Returns the completion state.
func (c *Completer) Completed() bool {
	return c.f.completed
}
// Completes the Completer with the given error and triggers al registered error handlers. Panics if the Completer is already
// complete.
func (c *Completer) CompleteError(err error) {
	c.f.completeError(err)
}