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
type Reactor[T any] struct {
	*sync.Mutex
	evtIn             *StreamController[Event[T]]
	shutdownCompleter *Completer[Data]
	shutdownReason    Data
	eventRegister     map[interface{}]Subscriber[T]
}

// NewReactor creates a new Reactor.
func NewReactor[T any]() (r *Reactor[T]) {

	r = &Reactor[T]{
		Mutex:         new(sync.Mutex),
		evtIn:         NewStreamController[Event[T]](),
		eventRegister: map[interface{}]Subscriber[T]{},
	}
	r.shutdownCompleter = r.evtIn.Stream().Listen(r.react)
	return
}

// OnShutdown registers a custom handler for the shutdown event.
func (r *Reactor[T]) OnShutdown(s Subscriber[Data]) {
	r.React(ShutdownEvent{}, func(_ T) { s(r.shutdownReason) })
}

// Shutdown shuts down the reactor, cancelling all go routines and stream subscriptions. The error is to fullfill the `Shutdowner` interface and will always be nil.
func (r *Reactor[T]) Shutdown(reason Data) (err error) {
	r.shutdownReason = reason
	var res T
	r.Fire(ShutdownEvent{}, res)
	return nil
}

// ShutdownFuture returns a future which gets completed after the reactor shut down.
func (r *Reactor[T]) ShutdownFuture() *Future[Data] {
	return r.shutdownCompleter.Future()
}

func (r *Reactor[T]) shutdown() {
	r.shutdownCompleter.Complete(nil)
}

// Fire triggers an event, invoking asynchronly the registered subscriber, if any. Events are guaranteed to be handled in the order of arrival.
func (r *Reactor[T]) Fire(classifier interface{}, data T) {
	if !r.shutdownCompleter.Completed() {
		r.evtIn.Add(Event[T]{classifier, data})
	}
}

// FireIn fires the given event after the given duration. FireIn can not be canceled.
func (r *Reactor[T]) FireIn(classifier interface{}, data T, duration time.Duration) {
	go r.fireIn(classifier, data, duration)
}

func (r *Reactor[T]) fireIn(classifier interface{}, data T, d time.Duration) {
	time.Sleep(d)
	if r.shutdownCompleter.Completed() {
		return
	}
	r.evtIn.Add(Event[T]{classifier, data})
}

// FireEvery fires the given event repeatedly. FireEvery can not be canceled and will run until the reactor is shut down.
func (r *Reactor[T]) FireEvery(classifier interface{}, data T, interval time.Duration) {
	go r.fireEvery(classifier, data, interval)
}

func (r *Reactor[T]) fireEvery(classifier interface{}, data T, d time.Duration) {
	for {
		time.Sleep(d)
		if r.shutdownCompleter.Completed() {
			return
		}
		r.evtIn.Add(Event[T]{classifier, data})
	}
}

// React registers a Subscriber as handler for a given event classier. Previously registered handlers for for the given classifier will be overwritten!
func (r *Reactor[T]) React(classifier interface{}, handler Subscriber[T]) {
	r.Lock()
	defer r.Unlock()
	r.eventRegister[classifier] = handler
}

func (r *Reactor[T]) react(evt Event[T]) {
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
func (r *Reactor[T]) AddStream(classifier interface{}, s *Stream[T]) {
	s.Listen(r.createEventFromStream(classifier)).CompleteOnFuture(r.shutdownCompleter.Future())
}

func (r *Reactor[T]) createEventFromStream(classifier interface{}) Subscriber[T] {
	return func(d T) {
		r.evtIn.Add(Event[T]{classifier, d})
	}
}

// AddFuture creates an event with given classifier, which will be fired when the given future completes. The event will not be triggered on error comletion.
func (r *Reactor[T]) AddFuture(classifier interface{}, f *Future[T]) {
	f.Then(r.createEventFromFuture(classifier))
}

func (r *Reactor[T]) createEventFromFuture(classifier interface{}) CompletionHandler[T] {
	return func(d T) {
		r.Fire(classifier, d)
	}
}

// AddObservable fires an event with the given classifier whenever the observable is changed.
func (r *Reactor[T]) AddObservable(classifier interface{}, o *Observable[T]) {
	o.OnChange(r.createEventFromObservable(classifier))
}

func (r *Reactor[T]) createEventFromObservable(classifier interface{}) Subscriber[T] {
	return func(d T) {
		r.Fire(classifier, d)
	}
}

// CollectEvent register the given Collectors Add method as eventhandler for the given classifier.
func (r *Reactor[T]) CollectEvent(classifier interface{}, c *Collector[T]) {
	r.React(classifier, c.Add)
}

// CatchCtrlC starts a goroutine, which initializes reactor shutdown when os.Interrupt is received.
func (r *Reactor[T]) CatchCtrlC() {
	go r.waitForCtrlC()
}

func (r *Reactor[T]) waitForCtrlC() {
	c := make(chan os.Signal, 1)
	defer close(c)
	signal.Notify(c, os.Interrupt)
	select {
	case s := <-c:
		r.Shutdown(s)
	case <-r.shutdownCompleter.Future().AsChan():
	}
}
