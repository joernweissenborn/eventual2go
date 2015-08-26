package eventual2go
import (
	"sync"
)

type Reactor struct {

	m *sync.Mutex

	evtIn StreamController
	evtC chan Data

	shutdown *Future

	eventRegister map[string]Subscriber

}

func NewReactor() (r *Reactor) {

	r = new(Reactor)

	r.m = new(sync.Mutex)

	r.evtIn = NewStreamController()
	r.evtC = r.evtIn.AsChan()

	r.shutdown = NewFuture()

	r.eventRegister = map[string]Subscriber{}

	go r.react()

	return
}

func (r *Reactor) Shutdown() {
	r.shutdown.Complete(nil)
	r.evtIn.Close()
}

func (r *Reactor) Fire(name string, data Data) {
	r.evtIn.Add(Event{name,data})
}

func (r *Reactor) React(name string, handler Subscriber) {
	r.m.Lock()
	defer r.m.Unlock()
	r.eventRegister[name] = handler
}

func (r *Reactor) AddStream(name string, s Stream) {
	s.Listen(r.createEventFromStream(name)).CloseOnFuture(r.shutdown)
}

func (r *Reactor) createEventFromStream(name string) Subscriber {
	return func(d Data) {
		r.evtIn.Add(Event{name,d})
	}
}

func (r *Reactor) AddFuture(name string, f *Future) {
	f.Then(r.createEventFromFuture(name))
}

func (r *Reactor) createEventFromFuture(name string) CompletionHandler {
	return func(d Data) Data{
		if !r.shutdown.IsComplete(){
			r.evtIn.Add(Event{name,d})
		}
		return nil
	}
}

func (r *Reactor) AddFutureError(name string, f *Future) {
	f.Err(r.createEventFromFutureError(name))
}

func (r *Reactor) createEventFromFutureError(name string) ErrorHandler {
	return func(e error) (Data,error){
		if !r.shutdown.IsComplete() {
			r.evtIn.Add(Event{name, e})
		}
		return nil,nil
	}
}

func (r *Reactor) react()  {
	for d := range r.evtC{
		evt := d.(Event)
		r.m.Lock()
		if h, f := r.eventRegister[evt.Name]; f {
			h(evt.Data)
		}
		r.m.Unlock()
	}
}
