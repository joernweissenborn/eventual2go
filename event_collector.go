package eventual2go

// EventCollector is a collector specific for `Event`.
type eventCollector struct {
	*Collector
}

func newEventCollector() (c *eventCollector) {
	c = &eventCollector{
		Collector: NewCollector(),
	}
	return
}

func (c *eventCollector) Add(e Event) {
	c.Collector.Add(e)
}

func (c *eventCollector) AddFuture(f *eventFuture) {
	c.Collector.AddFuture(f.Future)
}

func (c *eventCollector) AddStream(s *eventStream) {
	c.Collector.AddStream(s.Stream)
}

func (c *eventCollector) Get() Event {
	return c.Collector.Get().(Event)
}

func (c *eventCollector) Preview() Event {
	return c.Collector.Preview().(Event)
}
