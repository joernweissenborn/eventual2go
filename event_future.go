package eventual2go

type eventCompleter struct {
	*Completer
}

func newEventCompleter() *eventCompleter {
	return &eventCompleter{NewCompleter()}
}

func (c *eventCompleter) Complete(d Event) {
	c.Completer.Complete(d)
}

func (c *eventCompleter) Future() *eventFuture {
	return &eventFuture{c.Completer.Future()}
}

type eventFuture struct {
	*Future
}

type eventCompletionHandler func(Event) Event

func (ch eventCompletionHandler) toCompletionHandler() CompletionHandler {
	return func(d Data) Data {
		return ch(d.(Event))
	}
}

func (f *eventFuture) Then(ch eventCompletionHandler) *eventFuture {
	return &eventFuture{f.Future.Then(ch.toCompletionHandler())}
}

func (f *eventFuture) AsChan() chan Event {
	c := make(chan Event, 1)
	cmpl := func(d chan Event) eventCompletionHandler {
		return func(e Event) Event {
			d <- e
			close(d)
			return e
		}
	}
	ecmpl := func(d chan Event) ErrorHandler {
		return func(error) (Data, error) {
			close(d)
			return nil, nil
		}
	}
	f.Then(cmpl(c))
	f.Err(ecmpl(c))
	return c
}
