package eventual2go

import "sync"

// Observable is represents a value, which can be updated in a threadsafe and change order preserving manner.
//
// Subscribers can get informed by changes through either callbacks or channels.
type Observable struct {
	m      *sync.RWMutex // protects only the value
	change *StreamController
	value  Data
}

// NewObservable creates a new Observable with an initial value.
func NewObservable(initial Data) (o *Observable) {
	o = &Observable{
		m:      &sync.RWMutex{},
		change: NewStreamController(),
		value:  initial,
	}
	o.change.Stream().Listen(o.onChange)
	return
}

// Value returns the current value of the observable. Threadsafe.
func (o *Observable) Value() (value Data) {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.value
}

// Change changes the value of the observable
func (o *Observable) Change(value Data) {
	o.change.Add(value)

}

// OnChange registers a subscriber for change events.
func (o *Observable) OnChange(subscriber Subscriber) (cancel *Completer) {
	return o.change.Stream().Listen(subscriber)
}

func (o *Observable) onChange(value Data) {
	o.m.Lock()
	defer o.m.Unlock()
	o.value = value
}

// Stream returns a stream of change events.
func (o *Observable) Stream() (stream *Stream) {
	return o.change.Stream()
}

// AsChan returns a channel on which changes get send.
func (o *Observable) AsChan() (c chan Data, cancel *Completer) {
	return o.change.Stream().AsChan()
}

// NextChange returns a Future which gets completed with the next change.
func (o *Observable) NextChange() (f *Future) {
	return o.change.Stream().First()
}
