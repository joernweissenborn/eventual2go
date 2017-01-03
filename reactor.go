package eventual2go

import (
	"os"
	"os/signal"
	"sync"
	"time"
)

type ShutdownEvent struct{}

type Reactor struct {
	*sync.Mutex
	evtIn             *EventStreamController
	shutdownCompleter *Completer
	eventRegister     map[interface{}]Subscriber
}

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

func (r *Reactor) OnShutdown(s Subscriber) {
	r.React(ShutdownEvent{}, s)
}

func (r *Reactor) Shutdown(d Data) {
	r.Fire(ShutdownEvent{}, d)
}

func (r *Reactor) shutdown() {
	r.shutdownCompleter.Complete(nil)
}

func (r *Reactor) Fire(classifer interface{}, data Data) {
	r.evtIn.Add(Event{classifer, data})
}

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

func (r *Reactor) React(classifer interface{}, handler Subscriber) {
	r.Lock()
	defer r.Unlock()
	r.eventRegister[classifer] = handler
}

func (r *Reactor) AddStream(classifer interface{}, s *Stream) {
	s.Listen(r.createEventFromStream(classifer)).CompleteOnFuture(r.shutdownCompleter.Future())
}

func (r *Reactor) createEventFromStream(classifer interface{}) Subscriber {
	return func(d Data) {
		r.evtIn.Add(Event{classifer, d})
	}
}

func (r *Reactor) AddFuture(classifer interface{}, f *Future) {
	f.Then(r.createEventFromFuture(classifer))
}

func (r *Reactor) CollectEvent(classifer interface{}, c *Collector) {
	r.React(classifer, c.Add)
}

func (r *Reactor) createEventFromFuture(classifer interface{}) CompletionHandler {
	return func(d Data) Data {
		if !r.shutdownCompleter.Completed() {
			r.evtIn.Add(Event{classifer, d})
		}
		return nil
	}
}

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

func (r *Reactor) react(ec chan Event) {

	for evt := range ec {
		r.Lock()
		if h, f := r.eventRegister[evt.Classifier]; f {
			h(evt.Data)
		}
		if evt.Classifier == (ShutdownEvent{}) {
			r.shutdown()
		}
		r.Unlock()
	}
}

func (r *Reactor) CatchCtrlC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go r.waitForInterrupt(c)
}

func (r *Reactor) waitForInterrupt(c chan os.Signal) {
	defer close(c)
	r.Shutdown(<-c)
}
