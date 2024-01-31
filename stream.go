package eventual2go

import (
	"sync"
)

// A Stream can be consumed or new streams be derived by registering handler functions.
type Stream[T any] struct {
	m     *sync.Mutex
	next  *Future[*streamEvent[T]]
	close *Completer[Data]
}

// NewStream returns a new stream. Data can not be added to a Stream manually, use a StreamController instead.
func newStream[T any](next *Future[*streamEvent[T]]) (s *Stream[T]) {
	s = &Stream[T]{
		m:     &sync.Mutex{},
		next:  next,
		close: NewCompleter[Data](),
	}
	return
}

func (s *Stream[T]) updateNext(next *Future[*streamEvent[T]]) {
	s.m.Lock()
	defer s.m.Unlock()
	s.next = next
}

// Close closes the Stream and its assigned StreamController.
func (s *Stream[T]) Close() {
	s.close.Complete(true)
}

// Closed returns a Future which completes upon closing of the Stream.
func (s *Stream[T]) Closed() (f *Future[Data]) {
	return s.close.Future()
}

// CloseOnFuture closes the Stream upon completion of Future.
func (s *Stream[T]) CloseOnFuture(f *Future[Data]) {
	s.close.CompleteOnFuture(f)
}

// Listen registers a subscriber. Returns a Completer, which can be used to terminate the subcription.
func (s *Stream[T]) Listen(sr Subscriber[T]) (stop *Completer[Data]) {
	stop = NewCompleter[Data]()
	s.m.Lock()
	defer s.m.Unlock()
	s.next.Then(listen(sr, stop.Future(), true))
	return
}

// ListenNonBlocking is the same as Listen, but the subscriber is not blocking the subcription.
func (s *Stream[T]) ListenNonBlocking(sr Subscriber[T]) (stop *Completer[Data]) {
	stop = NewCompleter[Data]()
	s.m.Lock()
	defer s.m.Unlock()
	s.next.Then(listen(sr, stop.Future(), false))
	return
}

func listen[T any](sr Subscriber[T], stop *Future[Data], block bool) CompletionHandler[*streamEvent[T]] {
	return func(evt *streamEvent[T]) {
		if !stop.Completed() {
			if block {
				sr(evt.data)
			} else {
				go sr(evt.data)
			}
			evt.next.Then(listen(sr, stop, block))
		}
	}
}

// Derive creates a derived stream from a DeriveSubscriber. Mainly used internally.
func DeriveStream[T, V any](s *Stream[T], dsr DeriveSubscriber[T, V]) (ds *Stream[V]) {
	sc := NewStreamController[V]()
	ds = sc.Stream()
	s.m.Lock()
	defer s.m.Unlock()
	s.next.Then(listen(derive(sc, dsr), ds.close.Future(), true))
	return
}

func derive[T, V any](sc *StreamController[V], dsr DeriveSubscriber[T, V]) Subscriber[T] {
	return func(d T) {
		dsr(sc, d)
	}
}

// Transform registers a Transformer function and returns the transformed stream.
func TransformStream[T, V any](s *Stream[T], t Transformer[T, V]) (ts *Stream[V]) {
	ts = DeriveStream(s, transform(t))
	return
}

func transform[T, V any](t Transformer[T, V]) DeriveSubscriber[T, V] {
	return func(sc *StreamController[V], d T) {
		sc.Add(t(d))
	}
}

// TransformConditional registers a TransformerConditional function and returns the transformed stream.
func TransformStreamConditional[T, V any](s *Stream[T], t TransformerConditional[T, V]) (ts *Stream[V]) {
	ts = DeriveStream(s, transformConditional(t))
	return
}

func transformConditional[T, V any](t TransformerConditional[T, V]) DeriveSubscriber[T, V] {
	return func(sc *StreamController[V], d T) {
		if transformed, ok := t(d); ok {
			sc.Add(transformed)
		}
	}
}

// Where registers a Filter function and returns the filtered stream. Elements will be added if the Filter returns TRUE.
func (s *Stream[T]) Where(f ...Filter[T]) (fs *Stream[T]) {
	fs = DeriveStream(s, filter(f))
	return
}

func filter[T any](f []Filter[T]) DeriveSubscriber[T, T] {
	return func(sc *StreamController[T], d T) {
		for _, ff := range f {
			if !ff(d) {
				return
			}
		}
		sc.Add(d)
	}
}

// TransformWhere transforms only filtered elements.
func TransformStreamWhere[T, V any](s *Stream[T], t Transformer[T, V], f ...Filter[T]) (tws *Stream[V]) {
	fs := s.Where(f...)
	tws = TransformStream(fs, t)
	fs.CloseOnFuture(tws.Closed())
	return
}

// WhereNot registers a Filter function and returns the filtered stream. Elements will be added if the Filter returns FALSE.
func (s *Stream[T]) WhereNot(f ...Filter[T]) (fs *Stream[T]) {
	fs = DeriveStream(s,filterNot(f))
	return
}

func filterNot[T any](f []Filter[T]) DeriveSubscriber[T, T] {
	return func(sc *StreamController[T], d T) {
		for _, ff := range f {
			if ff(d) {
				return
			}
		}
		sc.Add(d)
	}
}

// First returns a future that will be completed with the first element added to the stream.
func (s *Stream[T]) First() (f *Future[T]) {
	c := NewCompleter[T]()
	f = c.Future()
	s.m.Lock()
	defer s.m.Unlock()
	s.next.Then(first(c))
	return
}

func first[T any](c *Completer[T]) CompletionHandler[*streamEvent[T]] {
	return func(evt *streamEvent[T]) {
		c.Complete(evt.data)
	}
}

// FirstWhere returns a future that will be completed with the first element added to the stream where filter returns TRUE.
func (s *Stream[T]) FirstWhere(f ...Filter[T]) (fw *Future[T]) {
	c := NewCompleter[T]()
	fw = c.Future()
	w := s.Where(f...)
	w.Listen(closeOnFirst(c, w))
	return
}

func closeOnFirst[T any](c *Completer[T], s *Stream[T]) Subscriber[T] {
	return func(d T) {
		c.Complete(d)
		s.Close()
	}
}

// FirstWhereNot returns a future that will be completed with the first element added to the stream where filter returns FALSE.
func (s *Stream[T]) FirstWhereNot(f ...Filter[T]) (fw *Future[T]) {
	c := NewCompleter[T]()
	fw = c.Future()
	w := s.WhereNot(f...)
	w.Listen(closeOnFirst(c, w))
	return
}

// Split returns a stream with all elements where the filter returns TRUE and one where the filter returns FALSE.
func (s *Stream[T]) Split(f Filter[T]) (ts *Stream[T], fs *Stream[T]) {
	return s.Where(f), s.WhereNot(f)
}

// AsChan returns a channel where all items will be pushed. Note items while be queued in a fifo since the stream must
// not block.
func (s *Stream[T]) AsChan() (c chan T, stop *Completer[Data]) {
	c = make(chan T)
	stop = s.Listen(pipeToChan(c))
	stop.Future().Then(closeChan(c))
	return
}

func pipeToChan[T any](c chan T) Subscriber[T] {
	return func(d T) {
		c <- d
	}
}
func closeChan[T any](c chan T) CompletionHandler[Data] {
	return func(d Data)  {
		close(c)
	}
}
