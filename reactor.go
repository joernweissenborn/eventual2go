package eventual2go

import (
	"os"
	"os/signal"
	"sync"
	"time"
)

type Reactor struct {
	m *sync.Mutex

	evtIn *EventStreamController

	shutdown *Completer

	eventRegister map[string]Subscriber //TODO: the suscriptions itself can be used
}

func NewReactor() (r *Reactor) {

	r = &Reactor{
		m:             new(sync.Mutex),
		evtIn:         NewEventStreamController(),
		shutdown:      NewCompleter(),
		eventRegister: map[string]Subscriber{},
	}

	go r.react(r.evtIn.Stream().AsChan())

	return
}

func (r *Reactor) Shutdown() {
	//TODO: shut be a special event
	r.shutdown.Complete(nil)
	r.evtIn.Close()
}

func (r *Reactor) Fire(name string, data Data) {
	r.evtIn.Add(Event{name, data})
}

func (r *Reactor) FireEvery(name string, data Data, interval time.Duration) {
	go r.fireEvery(name, data, interval)
}

func (r *Reactor) fireEvery(name string, data Data, d time.Duration) {
	c := true
	for c {
		time.Sleep(d)
		r.evtIn.Add(Event{name, data})
		c = !r.shutdown.Completed()
	}
}

func (r *Reactor) React(name string, handler Subscriber) {
	r.m.Lock()
	defer r.m.Unlock()
	r.eventRegister[name] = handler
}

func (r *Reactor) AddStream(name string, s *Stream) {
	s.Listen(r.createEventFromStream(name)).CloseOnFuture(r.shutdown.Future())
}

func (r *Reactor) createEventFromStream(name string) Subscriber {
	return func(d Data) {
		r.evtIn.Add(Event{name, d})
	}
}

func (r *Reactor) AddFuture(name string, f *Future) {
	f.Then(r.createEventFromFuture(name))
}

func (r *Reactor) createEventFromFuture(name string) CompletionHandler {
	return func(d Data) Data {
		if !r.shutdown.Completed() {
			r.evtIn.Add(Event{name, d})
		}
		return nil
	}
}

func (r *Reactor) AddFutureError(name string, f *Future) {
	f.Err(r.createEventFromFutureError(name))
}

func (r *Reactor) CatchCtrlC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go r.waitForInterrupt(c)
}

func (r *Reactor) waitForInterrupt(c chan os.Signal) {
	defer close(c)
	<-c
	r.Shutdown()
}

func (r *Reactor) createEventFromFutureError(name string) ErrorHandler {
	return func(e error) (Data, error) {
		if !r.shutdown.Completed() {
			r.evtIn.Add(Event{name, e})
		}
		return nil, nil
	}
}

func (r *Reactor) react(ec chan Event) {

	for evt := range ec {
		r.m.Lock()
		if h, f := r.eventRegister[evt.Name]; f {
			h(evt.Data)
		}
		r.m.Unlock()
	}
}
