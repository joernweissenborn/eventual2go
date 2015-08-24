package eventual2go
import "sync"

type Reactor struct {

	m *sync.Mutex

	evtIn chan Event

	shutdown *Future

	eventRegister map[string]Subscriber

}

func NewReactor() (r Reactor) {

	r.m = new(sync.Mutex)

	r.evtIn = make(chan Event)

	r.shutdown = NewFuture()

	r.eventRegister = map[string]Subscriber{}

	go r.react()

	return
}

func (r Reactor) Shutdown() {
	r.shutdown.Complete(nil)
	close(r.evtIn)
}

func (r Reactor) Fire(name string, data Data) {
	r.evtIn <- Event{name,data}
}

func (r Reactor) React(name string, handler Subscriber) {
	r.m.Lock()
	defer r.m.Unlock()
	r.eventRegister[name] = handler
}

func (r Reactor) AddStream(name string, s Stream) {
	s.Listen(r.createEventFromStream(name)).CloseOnFuture(r.shutdown)
}

func (r Reactor) createEventFromStream(name string) Subscriber {
	return func(d Data) {
		r.evtIn <- Event{name,d}
	}
}

func (r Reactor) AddFuture(name string, f *Future) {
	f.Then(r.createEventFromFuture(name))
}

func (r Reactor) createEventFromFuture(name string) CompletionHandler {
	return func(d Data) Data{
		if !r.shutdown.IsComplete(){
			r.evtIn <- Event{name,d}
		}
		return nil
	}
}

func (r Reactor) AddFutureError(name string, f *Future) {
	f.Err(r.createEventFromFutureError(name))
}

func (r Reactor) createEventFromFutureError(name string) ErrorHandler {
	return func(e error) (Data,error){
		if !r.shutdown.IsComplete() {
			r.evtIn <- Event{name, e}
		}
		return nil,nil
	}
}

func (r Reactor) react()  {
	for evt := range r.evtIn{
		r.m.Lock()
		if h, f := r.eventRegister[evt.Name]; f {
			h(evt.Data)
		}
		r.m.Unlock()
	}
}
