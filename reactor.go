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
		Mutex:             new(sync.Mutex),
		evtIn:             NewEventStreamController(),
		shutdownCompleter: NewCompleter(),
		eventRegister:     map[interface{}]Subscriber{},
	}
	evtC, stop := r.evtIn.Stream().AsChan()
	stop.CompleteOnFuture(r.shutdownCompleter.Future())
	go r.react(evtC)

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

func (r *Reactor) shutdown() {
	r.shutdownCompleter.Complete(nil)
}

// Fire fires an event, invoking asynchronly the Subscriber registered, if any. Events are guaranteed to be handled in the order of arrival.
func (r *Reactor) Fire(classifer interface{}, data Data) {
	r.evtIn.Add(Event{classifer, data})
}

// FireEvery fires the given event repeatedly. FireEvery can not be canceled and will run until the reactor is shut down.
func (r *Reactor) FireEvery(classifer interface{}, data Data, interval time.Duration) {
	go r.fireEvery(classifer, data, interval)
}

func (r *Reactor) fireEvery(classifer interface{}, data Data, d time.Duration) {
	for {
		time.Sleep(d)
		if r.shutdownCompleter.Completed() {
			return
		}
		r.evtIn.Add(Event{classifer, data})
	}
}

// React registers a Subscriber as handler for a given event classier. Previously registered handlers for vlassifier will be overwritten!
func (r *Reactor) React(classifer interface{}, handler Subscriber) {
	r.Lock()
	defer r.Unlock()
	r.eventRegister[classifer] = handler
}

func (r *Reactor) react(ec chan Event) {

	for evt := range ec {
		r.Lock()
		if h, f := r.eventRegister[evt.Classifier]; f {
			h(evt.Data)
		} else if evt.Classifier == (ShutdownEvent{}) {
			r.shutdown()
		}
		r.Unlock()
	}
}

// AddStream subscribes to a Stream, firing an event with the given classifier for every new element in the stream.
func (r *Reactor) AddStream(classifer interface{}, s *Stream) {
	s.Listen(r.createEventFromStream(classifer)).CompleteOnFuture(r.shutdownCompleter.Future())
}

func (r *Reactor) createEventFromStream(classifer interface{}) Subscriber {
	return func(d Data) {
		r.evtIn.Add(Event{classifer, d})
	}
}

// AddFuture adds a handler for  completion of the given Future which fires an event with the given classifier upon future completion.
func (r *Reactor) AddFuture(classifer interface{}, f *Future) {
	f.Then(r.createEventFromFuture(classifer))
}

func (r *Reactor) createEventFromFuture(classifer interface{}) CompletionHandler {
	return func(d Data) Data {
		if !r.shutdownCompleter.Completed() {
			r.evtIn.Add(Event{classifer, d})
		}
		return nil
	}
}

//AddFutureError acts the same as AddFuture, but registers a handler for the future error.
func (r *Reactor) AddFutureError(classifer interface{}, f *Future) {
	f.Err(r.createEventFromFutureError(classifer))
}

func (r *Reactor) createEventFromFutureError(classifer interface{}) ErrorHandler {
	return func(e error) (Data, error) {
		if !r.shutdownCompleter.Completed() {
			r.evtIn.Add(Event{classifer, e})
		}
		return nil, nil
	}
}

// CollectEvent register the given Collectors Add method as eventhandler for the given classifier.
func (r *Reactor) CollectEvent(classifer interface{}, c *Collector) {
	r.React(classifer, c.Add)
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
