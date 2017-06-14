package eventual2go

import "sync"

// A Stream can be consumed or new streams be derived by registering handler functions.
type Stream struct {
	m     *sync.Mutex
	next  *Future
	close *Completer
}

// NewStream returns a new stream. Data can not be added to a Stream manually, use a StreamController instead.
func newStream(next *Future) (s *Stream) {
	s = &Stream{
		m:     &sync.Mutex{},
		next:  next,
		close: NewCompleter(),
	}
	return
}

func (s *Stream)updateNext(next *Future) {
	s.m.Lock()
	defer s.m.Unlock()
	s.next = next
}

// Close closes the Stream and its assigned StreamController.
func (s *Stream) Close() {
	s.close.Complete(nil)
}

// Closed returns a Future which completes upon closing of the Stream.
func (s *Stream) Closed() (f *Future){
	return s.close.Future()
}

// CloseOnFuture closes the Stream upon completion of Future.
func (s *Stream) CloseOnFuture(f *Future) {
	s.close.CompleteOnFuture(f)
}

// Listen registers a subscriber. Returns a Completer, which can be used to terminate the subcription.
func (s *Stream) Listen(sr Subscriber) (stop *Completer) {
	stop = NewCompleter()
	s.m.Lock()
	defer s.m.Unlock()
	s.next.Then(listen(sr, stop.Future()))
	return
}

func listen(sr Subscriber, stop *Future) CompletionHandler {
	return func(d Data) Data {
		if !stop.Completed() {
			evt := d.(*streamEvent)
			sr(evt.data)
			evt.next.Then(listen(sr, stop))
		}
		return nil
	}
}

// Derive creates a derived stream from a DeriveSubscriber. Mainly used internally.
func (s *Stream) Derive(dsr DeriveSubscriber) (ds *Stream) {
	sc := NewStreamController()
	ds = sc.Stream()
	s.m.Lock()
	defer s.m.Unlock()
	s.next.Then(listen(derive(sc, dsr), ds.close.Future()))
	return
}

func derive(sc *StreamController, dsr DeriveSubscriber) Subscriber {
	return func(d Data) {
		dsr(sc, d)
	}
}

// Transform registers a Transformer function and returns the transformed stream.
func (s *Stream) Transform(t Transformer) (ts *Stream) {
	ts = s.Derive(transform(t))
	return
}

func transform(t Transformer) DeriveSubscriber {
	return func(sc *StreamController, d Data) {
		sc.Add(t(d))
	}
}

// Where registers a Filter function and returns the filtered stream. Elements will be added if the Filter returns TRUE.
func (s *Stream) Where(f ...Filter) (fs *Stream) {
	fs = s.Derive(filter(f))
	return
}

func filter(f []Filter) DeriveSubscriber {
	return func(sc *StreamController, d Data) {
		for _, ff := range f {
			if !ff(d) {
				return
			}
		}
		sc.Add(d)
	}
}

// TransformWhere transforms only filtered elements.
func (s *Stream) TransformWhere(t Transformer, f ...Filter) (tws *Stream) {
	fs := s.Where(f...)
	tws = fs.Transform(t)
	fs.CloseOnFuture(tws.Closed())
	return
}
// WhereNot registers a Filter function and returns the filtered stream. Elements will be added if the Filter returns FALSE.
func (s *Stream) WhereNot(f ...Filter) (fs *Stream) {
	fs = s.Derive(filterNot(f))
	return
}

func filterNot(f []Filter) DeriveSubscriber {
	return func(sc *StreamController, d Data) {
		for _, ff := range f {
			if ff(d) {
				return
			}
		}
		sc.Add(d)
	}
}

// First returns a future that will be completed with the first element added to the stream.
func (s *Stream) First() (f *Future) {
	c := NewCompleter()
	f = c.Future()
	s.next.Then(first(c))
	return
}

func first(c *Completer) CompletionHandler {
	return func(d Data) Data {
		evt := d.(*streamEvent)
		c.Complete(evt.data)
		return nil
	}
}

// FirstWhere returns a future that will be completed with the first element added to the stream where filter returns TRUE.
func (s *Stream) FirstWhere(f ...Filter) (fw *Future) {
	c := NewCompleter()
	fw = c.Future()
	w := s.Where(f...)
	w.Listen(closeOnFirst(c, w))
	return
}

func closeOnFirst(c *Completer, s *Stream) Subscriber {
	return func(d Data) {
		c.Complete(d)
		s.Close()
	}
}

// FirstWhereNot returns a future that will be completed with the first element added to the stream where filter returns FALSE.
func (s *Stream) FirstWhereNot(f ...Filter) (fw *Future) {
	c := NewCompleter()
	fw = c.Future()
	w := s.WhereNot(f...)
	w.Listen(closeOnFirst(c, w))
	return
}

// Split returns a stream with all elements where the filter returns TRUE and one where the filter returns FALSE.
func (s *Stream) Split(f Filter) (ts *Stream, fs *Stream) {
	return s.Where(f), s.WhereNot(f)
}

// AsChan returns a channel where all items will be pushed. Note items while be queued in a fifo since the stream must
// not block.
func (s *Stream) AsChan() (c chan Data, stop *Completer) {
	c = make(chan Data)
	stop = s.Listen(pipeToChan(c))
	stop.Future().Then(closeChan(c))
	return
}

func pipeToChan(c chan Data) Subscriber {
	return func(d Data) {
		c <- d
	}
}
func closeChan(c chan Data) CompletionHandler {
	return func(d Data) Data {
		close(c)
		return nil
	}
}
