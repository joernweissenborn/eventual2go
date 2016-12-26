package eventual2go

type EventCompleter struct {
	*Completer
}

func NewEventCompleter() *EventCompleter {
	return &EventCompleter{NewCompleter()}
}

func (c *EventCompleter) Complete(d Event) {
	c.Completer.Complete(d)
}

func (c *EventCompleter) Future() *EventFuture {
	return &EventFuture{c.Completer.Future()}
}

type EventFuture struct {
	*Future
}

type EventCompletionHandler func(Event) Event

func (ch EventCompletionHandler) toCompletionHandler() CompletionHandler {
	return func(d Data) Data {
		return ch(d.(Event))
	}
}

func (f *EventFuture) Then(ch EventCompletionHandler) *EventFuture {
	return &EventFuture{f.Future.Then(ch.toCompletionHandler())}
}

func (f *EventFuture) AsChan() chan Event {
	c := make(chan Event, 1)
	cmpl := func(d chan Event) EventCompletionHandler {
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
