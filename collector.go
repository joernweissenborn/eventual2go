package eventual2go

// Collector is a data sink. Use it to collect events for later retrieval. All events are stored in historical order.
type Collector[T any] struct {
	r      *Reactor[T]
	pile   []T
	remove *Completer[Data]
}

type addEvent struct{}

// NewCollector creates a new Collector.
func NewCollector[T any]() (c *Collector[T]) {
	c = &Collector[T]{
		r:      NewReactor[T](),
		remove: NewCompleter[Data](),
	}
	c.r.React(addEvent{}, c.collect)
	return
}

func (c *Collector[T]) collect(d T) {
	c.pile = append(c.pile, d)
}

// Stop stops the collection on events.
func (c *Collector[T]) Stop() {
	c.r.Shutdown(nil)
	c.remove.Complete(nil)
}

// Stopped returns a future which completes when the reactor collected all events before the call to `Stop`.
func (c *Collector[T]) Stopped() *Future[Data] {
	return c.remove.Future()
}

// Add adds data to the collector.
func (c *Collector[T]) Add(d T) {
	c.r.Fire(addEvent{}, d)
}

// AddStream collects all events on `Stream`
func (c *Collector[T]) AddStream(s *Stream[T]) {
	s.Listen(c.Add).CompleteOnFuture(c.remove.Future())
}

// AddObservable collects all changes off an `Observable`
func (c *Collector[T]) AddObservable(o *Observable[T]) {
	o.OnChange(c.Add)
}

// AddFuture collects the result of a `T`
func (c *Collector[T]) AddFuture(f *Future[T]) {
	f.Then(c.collectFuture)
}

func (c *Collector[T]) collectFuture(d T){
	c.Add(d)
}

// Get retrieves the oldes data from the collecter and deletes it from it.
func (c *Collector[T]) Get() (d Data) {
	c.r.Lock()
	defer c.r.Unlock()
	if len(c.pile) != 0 {
		d = c.pile[0]
		c.pile = c.pile[1:]
	}
	return
}

// Preview retrieves the oldes data from the collecter without deleting it from it.
func (c *Collector[T]) Preview() (d Data) {
	c.r.Lock()
	defer c.r.Unlock()
	if len(c.pile) != 0 {
		d = c.pile[0]
	}
	return
}

// Empty returns true if at least one data element is stored in the collector.
func (c *Collector[T]) Empty() (e bool) {
	c.r.Lock()
	defer c.r.Unlock()
	e = len(c.pile) == 0
	return
}

// Size returns the number of events stored in the collector.
func (c *Collector[T]) Size() (n int) {
	c.r.Lock()
	defer c.r.Unlock()
	n = len(c.pile)
	return
}
