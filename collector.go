package eventual2go

type Collector struct {
	r      *Reactor
	pile   []Data
	remove *Completer
}

func NewCollector() (c *Collector) {

	c = &Collector{
		NewReactor(),
		[]Data{},
		NewCompleter(),
	}

	c.r.React("add", c.collect)
	c.r.React("get", c.get)
	c.r.React("empty", c.empty)

	return
}

func (c *Collector) Remove() {
	c.r.Shutdown()
	c.remove.Complete(nil)
}

func (c *Collector) Add(d Data) {
	c.r.Fire("add", d)
}

func (c *Collector) AddStream(s *Stream) {
	s.Listen(c.Add).CloseOnFuture(c.remove.Future())
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
	cd := make(chan Data)
	defer close(cd)
	c.r.Fire("get", cd)
	d = <-cd
	return
}

func (c *Collector) Empty() (e bool) {
	cd := make(chan bool)
	defer close(cd)
	c.r.Fire("empty", cd)
	e = <-cd
	return
}

func (c *Collector) collect(d Data) {

	c.pile = append(c.pile, d)

}

func (c *Collector) get(d Data) {

	cd := d.(chan Data)

	if len(c.pile) != 0 {
		cd <- c.pile[0]
		c.pile = c.pile[1:]
	} else {
		cd <- nil
	}

}

func (c *Collector) empty(d Data) {

	cd := d.(chan bool)

	cd <- len(c.pile) == 0

}
