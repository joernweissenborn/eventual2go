package eventual2go

type Collector struct {
	r      *Reactor
	pile   []Data
	remove *Completer
}

type addEvent struct{}

func NewCollector() (c *Collector) {

	c = &Collector{
		r:      NewReactor(),
		remove: NewCompleter(),
	}

	c.r.React(addEvent{}, c.collect)
	return
}

func (c *Collector) collect(d Data) {
	c.pile = append(c.pile, d)
}

func (c *Collector) Remove() {
	c.r.Shutdown(nil)
	c.remove.Complete(nil)
}

func (c *Collector) Removed() *Future {
	return c.remove.Future()
}

func (c *Collector) Add(d Data) {
	c.r.Fire(addEvent{}, d)
}

func (c *Collector) AddStream(s *Stream) {
	s.Listen(c.Add).CompleteOnFuture(c.remove.Future())
}

func (c *Collector) AddObservable(o *Observable) {
	o.OnChange(c.Add)
}

func (c *Collector) AddFuture(f *Future) {
	f.Then(c.collectFuture)
}

func (c *Collector) collectFuture(d Data) Data {
	c.Add(d)
	return nil
}

func (c *Collector) AddFutureError(f *Future) {
	f.Err(c.collectFutureErr)
}

func (c *Collector) collectFutureErr(e error) (Data, error) {
	c.Add(e)
	return nil, nil
}

func (c *Collector) Get() (d Data) {
	c.r.Lock()
	defer c.r.Unlock()
	if len(c.pile) != 0 {
		d = c.pile[0]
		c.pile = c.pile[1:]
	}
	return
}

func (c *Collector) Preview() (d Data) {
	c.r.Lock()
	defer c.r.Unlock()
	if len(c.pile) != 0 {
		d = c.pile[0]
	}
	return
}

func (c *Collector) Empty() (e bool) {
	c.r.Lock()
	defer c.r.Unlock()
	e = len(c.pile) == 0
	return
}
