package eventual2go

// Collector is a data sink. Use it to collect events for later retrieval. All events are stored in historical order.
type Collector struct {
	r      *Reactor
	pile   []Data
	remove *Completer[Data]
}

type addEvent struct{}

// NewCollector creates a new Collector.
func NewCollector() (c *Collector) {
	c = &Collector{
		r:      NewReactor(),
		remove: NewCompleter[Data](),
	}
	c.r.React(addEvent{}, c.collect)
	return
}

func (c *Collector) collect(d Data) {
	c.pile = append(c.pile, d)
}

// Stop stops the collection on events.
func (c *Collector) Stop() {
	c.r.Shutdown(nil)
	c.remove.Complete(nil)
}

// Stopped returns a future which completes when the reactor collected all events before the call to `Stop`.
func (c *Collector) Stopped() *Future[Data] {
	return c.remove.Future()
}

// Add adds data to the collector.
func (c *Collector) Add(d Data) {
	c.r.Fire(addEvent{}, d)
}

// AddStream collects all events on `Stream`
func (c *Collector) AddStream(s *Stream[Data]) {
	s.Listen(c.Add).CompleteOnFuture(c.remove.Future())
}

// AddObservable collects all changes off an `Observable`
func (c *Collector) AddObservable(o *Observable[Data]) {
	o.OnChange(c.Add)
}

// AddFuture collects the result of a `Future`
func (c *Collector) AddFuture(f *Future[Data]) {
	f.Then(c.collectFuture)
}

func (c *Collector) collectFuture(d Data){
	c.Add(d)
}

// AddFutureError collects the error of a `Future`
func (c *Collector) AddFutureError(f *Future[Data]) {
	f.Err(c.collectFutureErr)
}

func (c *Collector) collectFutureErr(e error)  {
	c.Add(e)
}

// Get retrieves the oldes data from the collecter and deletes it from it.
func (c *Collector) Get() (d Data) {
	c.r.Lock()
	defer c.r.Unlock()
	if len(c.pile) != 0 {
		d = c.pile[0]
		c.pile = c.pile[1:]
	}
	return
}

// Preview retrieves the oldes data from the collecter without deleting it from it.
func (c *Collector) Preview() (d Data) {
	c.r.Lock()
	defer c.r.Unlock()
	if len(c.pile) != 0 {
		d = c.pile[0]
	}
	return
}

// Empty returns true if at least one data element is stored in the collector.
func (c *Collector) Empty() (e bool) {
	c.r.Lock()
	defer c.r.Unlock()
	e = len(c.pile) == 0
	return
}

// Size returns the number of events stored in the collector.
func (c *Collector) Size() (n int) {
	c.r.Lock()
	defer c.r.Unlock()
	n = len(c.pile)
	return
}
