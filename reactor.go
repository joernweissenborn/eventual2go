package eventual2go

import (
	"os"
	"os/signal"
	"sync"
	"time"
)

// ShutdownEvent is triggering the reactor Shutdown when fired.
type ShutdownEvent struct{}

// Reactor is thread-safe event handler.
type Reactor struct {
	*sync.Mutex
	evtIn             *EventStreamController
	shutdownCompleter *Completer
	eventRegister     map[interface{}]Subscriber
}

// NewReactor creates a new Reactor.
func NewReactor() (r *Reactor) {

	r = &Reactor{
		Mutex:         new(sync.Mutex),
		evtIn:         NewEventStreamController(),
		eventRegister: map[interface{}]Subscriber{},
	}
	r.shutdownCompleter = r.evtIn.Stream().Listen(r.react)
	return
}

// OnShutdown registers a custom handler for the shutdown event.
func (r *Reactor) OnShutdown(s Subscriber) {
	r.React(ShutdownEvent{}, s)
}

// Shutdown shuts down the reactor, cancelling all go routines and stream subscriptions.
func (r *Reactor) Shutdown(d Data) {
	r.Fire(ShutdownEvent{}, d)
}

// ShutdownFuture returns a future which gets completed after the reactor shut down.
func (r *Reactor) ShutdownFuture() *Future {
	return r.shutdownCompleter.Future()
}

func (r *Reactor) shutdown() {
	r.shutdownCompleter.Complete(nil)
}

// Fire triggers an event, invoking asynchronly the registered subscriber, if any. Events are guaranteed to be handled in the order of arrival.
func (r *Reactor) Fire(classifier interface{}, data Data) {
	if !r.shutdownCompleter.Completed() {
		r.evtIn.Add(Event{classifier, data})
	}
}

// FireEvery fires the given event repeatedly. FireEvery can not be canceled and will run until the reactor is shut down.
func (r *Reactor) FireEvery(classifier interface{}, data Data, interval time.Duration) {
	go r.fireEvery(classifier, data, interval)
}

func (r *Reactor) fireEvery(classifier interface{}, data Data, d time.Duration) {
	for {
		time.Sleep(d)
		if r.shutdownCompleter.Completed() {
			return
		}
		r.evtIn.Add(Event{classifier, data})
	}
}

// React registers a Subscriber as handler for a given event classier. Previously registered handlers for for the given classifier will be overwritten!
func (r *Reactor) React(classifier interface{}, handler Subscriber) {
	r.Lock()
	defer r.Unlock()
	r.eventRegister[classifier] = handler
}

func (r *Reactor) react(evt Event) {
	r.Lock()
	defer r.Unlock()
	if h, f := r.eventRegister[evt.Classifier]; f {
		h(evt.Data)
	}
	if _, is := evt.Classifier.(ShutdownEvent); is {
		r.shutdown()
	}
}

// AddStream subscribes to a Stream, firing an event with the given classifier for every new element in the stream.
func (r *Reactor) AddStream(classifier interface{}, s *Stream) {
	s.Listen(r.createEventFromStream(classifier)).CompleteOnFuture(r.shutdownCompleter.Future())
}

func (r *Reactor) createEventFromStream(classifier interface{}) Subscriber {
	return func(d Data) {
		r.evtIn.Add(Event{classifier, d})
	}
}

// AddFuture creates an event with given classifier, which will be fired when the given future completes. The event will not be triggered on error comletion.
func (r *Reactor) AddFuture(classifier interface{}, f *Future) {
	f.Then(r.createEventFromFuture(classifier))
}

func (r *Reactor) createEventFromFuture(classifier interface{}) CompletionHandler {
	return func(d Data) Data {
		r.Fire(classifier, d)
		return nil
	}
}

//AddFutureError acts the same as AddFuture, but registers a handler for the future error.
func (r *Reactor) AddFutureError(classifier interface{}, f *Future) {
	f.Err(r.createEventFromFutureError(classifier))
}

func (r *Reactor) createEventFromFutureError(classifier interface{}) ErrorHandler {
	return func(e error) (Data, error) {
		r.Fire(classifier, e)
		return nil, nil
	}
}

//AddObservable fires an event with the given classifier whenever the observable is changed.
func (r *Reactor) AddObservable(classifier interface{}, o *Observable) {
	o.OnChange(r.createEventFromObservable(classifier))
}

func (r *Reactor) createEventFromObservable(classifier interface{}) Subscriber {
	return func(d Data) {
		r.Fire(classifier, d)
	}
}

// CollectEvent register the given Collectors Add method as eventhandler for the given classifier.
func (r *Reactor) CollectEvent(classifier interface{}, c *Collector) {
	r.React(classifier, c.Add)
}

// CatchCtrlC starts a goroutine, which initializes reactor shutdown when os.Interrupt is received.
func (r *Reactor) CatchCtrlC() {
	go r.waitForCtrlC()
}

func (r *Reactor) waitForCtrlC() {
	c := make(chan os.Signal, 1)
	defer close(c)
	signal.Notify(c, os.Interrupt)
	select {
	case s := <-c:
		r.Shutdown(s)
	case <-r.shutdownCompleter.Future().AsChan():
	}
}
