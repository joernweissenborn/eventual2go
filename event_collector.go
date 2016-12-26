package eventual2go

type EventCollector struct {
	*Collector
}

func NewEventCollector() (c *EventCollector) {
	c = &EventCollector{
		Collector: NewCollector(),
	}
	return
}

func (c *EventCollector) Add(e Event) {
	c.Collector.Add(e)
}

func (c *EventCollector) AddFuture(f *EventFuture) {
	c.Collector.AddFuture(f.Future)
}

func (c *EventCollector) AddStream(s *EventStream) {
	c.Collector.AddStream(s.Stream)
}

func (c *EventCollector) Get() Event {
	return c.Collector.Get().(Event)
}

func (c *EventCollector) Preview() Event {
	return c.Collector.Preview().(Event)
}
