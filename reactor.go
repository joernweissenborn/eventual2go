package eventual2go

import (
	"os"
	"os/signal"
	"sync"
	"time"
)

const (
	ShutdownEvent = "shutdown"
)

type Reactor struct {
	*sync.Mutex
	evtIn             *EventStreamController
	shutdownCompleter *Completer
	eventRegister     map[string]Subscriber
}

func NewReactor() (r *Reactor) {

	r = &Reactor{
		Mutex:             new(sync.Mutex),
		evtIn:             NewEventStreamController(),
		shutdownCompleter: NewCompleter(),
		eventRegister:     map[string]Subscriber{},
	}

	r.evtIn.Stream().FirstWhere(func(e Event) bool {
		return e.Name == ShutdownEvent
	}).Then(r.shutdown)
	
	r.evtIn.Stream().CloseOnFuture(r.shutdownCompleter.Future())

	go r.react(r.evtIn.Stream().AsChan())

	return
}

func (r *Reactor) OnShutdown(s Subscriber) {
	r.React(ShutdownEvent, s)
}

func (r *Reactor) Shutdown(d Data) {
	r.evtIn.Close()
	r.Fire(ShutdownEvent, d)
}

func (r *Reactor) shutdown(e Event) Event {
	r.shutdownCompleter.Complete(e.Data)
	return e
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
		c = !r.shutdownCompleter.Completed()
	}
}

func (r *Reactor) React(name string, handler Subscriber) {
	r.Lock()
	defer r.Unlock()
	r.eventRegister[name] = handler
}

func (r *Reactor) AddStream(name string, s *Stream) {
	s.Listen(r.createEventFromStream(name)).CloseOnFuture(r.shutdownCompleter.Future())
}

func (r *Reactor) createEventFromStream(name string) Subscriber {
	return func(d Data) {
		r.evtIn.Add(Event{name, d})
	}
}

func (r *Reactor) AddFuture(name string, f *Future) {
	f.Then(r.createEventFromFuture(name))
}

func (r *Reactor) CollectEvent(name string, c *Collector) {
	r.React(name,c.Add)
}

func (r *Reactor) createEventFromFuture(name string) CompletionHandler {
	return func(d Data) Data {
		if !r.shutdownCompleter.Completed() {
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
	r.Shutdown(<-c)
}

func (r *Reactor) createEventFromFutureError(name string) ErrorHandler {
	return func(e error) (Data, error) {
		if !r.shutdownCompleter.Completed() {
			r.evtIn.Add(Event{name, e})
		}
		return nil, nil
	}
}

func (r *Reactor) react(ec chan Event) {

	for evt := range ec {
		r.Lock()
		if h, f := r.eventRegister[evt.Name]; f {
			h(evt.Data)
		}
		r.Unlock()
	}
}
