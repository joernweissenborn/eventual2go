package eventual2go

import "sync"

// Observable is represents a value, which can be updated in a threadsafe and change order preserving manner.
//
// Subscribers can get informed by changes through either callbacks or channels.
type Observable[T any] struct {
	m      *sync.RWMutex // protects only the value
	change *StreamController[T]
	value  T
}

// NewObservable creates a new Observable with an initial value.
func NewObservable[T any](initial T) (o *Observable[T]) {
	o = &Observable[T]{
		m:      &sync.RWMutex{},
		change: NewStreamController[T](),
		value:  initial,
	}
	o.change.Stream().Listen(o.onChange)
	return
}

// Value returns the current value of the observable. Threadsafe.
func (o *Observable[T]) Value() (value T) {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.value
}

// Change changes the value of the observable
func (o *Observable[T]) Change(value T) {
	o.change.Add(value)

}

// OnChange registers a subscriber for change events.
func (o *Observable[T]) OnChange(subscriber Subscriber[T]) (cancel *Completer[Data]) {
	return o.change.Stream().Listen(subscriber)
}

func (o *Observable[T]) onChange(value T) {
	o.m.Lock()
	defer o.m.Unlock()
	o.value = value
}

// Stream returns a stream of change events.
func (o *Observable[T]) Stream() (stream *Stream[T]) {
	return o.change.Stream()
}

// AsChan returns a channel on which changes get send.
func (o *Observable[T]) AsChan() (c chan T, cancel *Completer[Data]) {
	return o.change.Stream().AsChan()
}

// NextChange returns a Future which gets completed with the next change.
func (o *Observable[T]) NextChange() (f *Future[T]) {
	return o.change.Stream().First()
}

// Derive returns a new Observable which value will be set by transform function everytime the source gets updated.
func DeriveObservable[T, V any](o *Observable[T], t Transformer[T, V]) (do *Observable[V], cancel *Completer[Data]) {
	do = NewObservable(t(o.Value()))
	cancel = o.OnChange(func(d T) {
		do.Change(t(d))
	})
	return
}
