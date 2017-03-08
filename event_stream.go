package eventual2go

type EventStreamController struct {
	*StreamController
}

func NewEventStreamController() *EventStreamController {
	return &EventStreamController{NewStreamController()}
}

func (sc *EventStreamController) Add(d Event) {
	sc.StreamController.Add(d)
}

func (sc *EventStreamController) Join(s *EventStream) {
	sc.StreamController.Join(s.Stream)
}

func (sc *EventStreamController) JoinFuture(f *EventFuture) {
	sc.StreamController.JoinFuture(f.Future)
}

func (sc *EventStreamController) Stream() *EventStream {
	return &EventStream{sc.StreamController.Stream()}
}

type EventStream struct {
	*Stream
}

type EventSuscriber func(Event)

func (l EventSuscriber) toSuscriber() Subscriber {
	return func(d Data) { l(d.(Event)) }
}

func (s *EventStream) Listen(ss EventSuscriber) (stop *Completer) {
	return s.Stream.Listen(ss.toSuscriber())
}

type EventFilter func(Event) bool

func (f EventFilter) toFilter() Filter {
	return func(d Data) bool { return f(d.(Event)) }
}

func (s *EventStream) Where(f EventFilter) *EventStream {
	return &EventStream{s.Stream.Where(f.toFilter())}
}

func (s *EventStream) WhereNot(f EventFilter) *EventStream {
	return &EventStream{s.Stream.WhereNot(f.toFilter())}
}

func (s *EventStream) First() *EventFuture {
	return &EventFuture{s.Stream.First()}
}

func (s *EventStream) FirstWhere(f EventFilter) *EventFuture {
	return &EventFuture{s.Stream.FirstWhere(f.toFilter())}
}

func (s *EventStream) FirstWhereNot(f EventFilter) *EventFuture {
	return &EventFuture{s.Stream.FirstWhereNot(f.toFilter())}
}

func (s *EventStream) AsChan() (c chan Event, stop *Completer) {
	c = make(chan Event)
	stop = s.Listen(pipeToEventChan(c))
	stop.Future().Then(closeEventChan(c))
	return
}

func pipeToEventChan(c chan Event) EventSuscriber {
	return func(d Event) {
		c <- d
	}
}

func closeEventChan(c chan Event) CompletionHandler {
	return func(d Data) Data {
		close(c)
		return nil
	}
}
