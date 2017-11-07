package eventual2go

type eventStreamController struct {
	*StreamController
}

func newEventStreamController() *eventStreamController {
	return &eventStreamController{NewStreamController()}
}

func (sc *eventStreamController) Add(d Event) {
	sc.StreamController.Add(d)
}

func (sc *eventStreamController) Join(s *eventStream) {
	sc.StreamController.Join(s.Stream)
}

func (sc *eventStreamController) JoinFuture(f *eventFuture) {
	sc.StreamController.JoinFuture(f.Future)
}

func (sc *eventStreamController) Stream() *eventStream {
	return &eventStream{sc.StreamController.Stream()}
}

type eventStream struct {
	*Stream
}

type eventSuscriber func(Event)

func (l eventSuscriber) toSuscriber() Subscriber {
	return func(d Data) { l(d.(Event)) }
}

func (s *eventStream) Listen(ss eventSuscriber) (stop *Completer) {
	return s.Stream.Listen(ss.toSuscriber())
}

type eventFilter func(Event) bool

func (f eventFilter) toFilter() Filter {
	return func(d Data) bool { return f(d.(Event)) }
}

func (s *eventStream) Where(f eventFilter) *eventStream {
	return &eventStream{s.Stream.Where(f.toFilter())}
}

func (s *eventStream) WhereNot(f eventFilter) *eventStream {
	return &eventStream{s.Stream.WhereNot(f.toFilter())}
}

func (s *eventStream) First() *eventFuture {
	return &eventFuture{s.Stream.First()}
}

func (s *eventStream) FirstWhere(f eventFilter) *eventFuture {
	return &eventFuture{s.Stream.FirstWhere(f.toFilter())}
}

func (s *eventStream) FirstWhereNot(f eventFilter) *eventFuture {
	return &eventFuture{s.Stream.FirstWhereNot(f.toFilter())}
}

func (s *eventStream) AsChan() (c chan Event, stop *Completer) {
	c = make(chan Event)
	stop = s.Listen(pipeToEventChan(c))
	stop.Future().Then(closeEventChan(c))
	return
}

func pipeToEventChan(c chan Event) eventSuscriber {
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
