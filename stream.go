package eventual2go

import "sync"

// A Stream can be consumed or new streams be derived by registering handler functions.
type Stream struct {
	m             *sync.Mutex
	subscriptions map[*Subscription]interface{}
	closed        *Completer
}

// NewStream returns a new stream. Data can not be added to a Stream manually, use a StreamController instead.
func NewStream() (s *Stream) {
	s = &Stream{
		m:             &sync.Mutex{},
		subscriptions: map[*Subscription]interface{}{},
		closed:        NewCompleter(),
	}
	return
}

func (s *Stream) add(d Data) {
	s.m.Lock()
	defer s.m.Unlock()
	for ss := range s.subscriptions {
		ss.add(d)
	}
}

func (s *Stream) subscribe(ss *Subscription) {
	s.m.Lock()
	defer s.m.Unlock()
	s.subscriptions[ss] = nil
	ss.canceled.Future().Then(s.cancelSubscription)
}

func (s *Stream) unsubscribe(ss *Subscription) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.subscriptions, ss)

}

// Close closes a Stream and cancels all subscriptions.
func (s *Stream) Close() {
	s.m.Lock()
	defer s.m.Unlock()
	for ss := range s.subscriptions {
		ss.Close()
	}
	s.closed.Complete(nil)
}

// Closed returns a Future which is completed when the stream is closed.
func (s *Stream) Closed() *Future {
	return s.closed.Future()
}

// CloseOnFuture Closes the stream when the given Future (error-)completes.
func (s *Stream) CloseOnFuture(f *Future) {
	f.Then(s.closeOnComplete)
	f.Err(s.closeOnCompleteError)
}

func (s *Stream) closeOnComplete(Data) Data {
	s.Close()
	return nil
}
func (s *Stream) closeOnCompleteError(error) (Data, error) {
	s.Close()
	return nil, nil
}

// Listen registers a subscriber. Returns Subscription, which can be used to terminate the subcription.
func (s *Stream) Listen(sr Subscriber) (ss *Subscription) {
	ss = NewSubscription(s, sr)
	s.subscribe(ss)
	return
}
func (s *Stream) cancelSubscription(d Data) Data {
	if !s.closed.Completed() {
		ss := d.(*Subscription)
		s.unsubscribe(ss)

	}
	return nil
}

// Transform registers a Transformer function and returns the transformed stream.
func (s *Stream) Transform(t Transformer) (ts *Stream) {
	ts = NewStream()
	s.Listen(transform(ts, t)).CloseOnFuture(ts.closed.Future())
	return
}

func transform(s *Stream, t Transformer) Subscriber {
	return func(d Data) {
		s.add(t(d))
	}
}

// Where registers a Filter function and returns the filtered stream. Elements will be added if the Filter returns TRUE.
func (s *Stream) Where(f ...Filter) (fs *Stream) {
	fs = NewStream()
	s.Listen(filter(fs, f)).CloseOnFuture(fs.closed.Future())
	return
}

func filter(s *Stream, f []Filter) Subscriber {
	return func(d Data) {
		for _,ff := range f {
			if !ff(d) {
				return
			}
		}
		s.add(d)
	}
}

// WhereNot registers a Filter function and returns the filtered stream. Elements will be added if the Filter returns FALSE.
func (s *Stream) WhereNot(f Filter) (fs *Stream) {
	fs = NewStream()
	s.Listen(filterNot(fs, f)).CloseOnFuture(fs.closed.Future())
	return
}

func filterNot(s *Stream, f Filter) Subscriber {
	return func(d Data) {
		if !f(d) {
			s.add(d)
		}
	}
}

// First returns a future that will be completed with the first element added to the stream.
func (s *Stream) First() (f *Future) {
	c := NewCompleter()
	f = c.Future()
	s.Listen(first(c)).CloseOnFuture(f)
	return
}

func first(c *Completer) Subscriber {
	return func(d Data) {
		if !c.Completed() {
			c.Complete(d)
		}
	}
}

// FirstWhere returns a future that will be completed with the first element added to the stream where filter returns TRUE.
func (s *Stream) FirstWhere(f ...Filter) *Future {
	return s.Where(f...).First()
}

// FirstWhereNot returns a future that will be completed with the first element added to the stream where filter returns FALSE.
func (s *Stream) FirstWhereNot(f Filter) *Future {
	return s.WhereNot(f).First()
}

// Split returns a stream with all elements where the filter returns TRUE and one where the filter returns FALSE.
func (s *Stream) Split(f Filter) (ts *Stream, fs *Stream) {
	return s.Where(f), s.WhereNot(f)
}

// AsChan returns a channel where all items will be pushed. Note items while be queued in a fifo since the stream must
// not block.
func (s *Stream) AsChan() (c chan Data) {
	c = make(chan Data)
	s.Listen(pipeToChan(c)).Closed().Then(closeChan(c))
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
