package eventual2go

import (
	"reflect"
)

// A Stream can be consumed or new streams be derived by registering handler functions.
type Stream struct {
	add_subscription    chan *Subscription
	remove_subscription chan *Subscription
	data_in             chan Data

	subscriptions []*Subscription
	closed        *Completer

	stop chan struct{}
}

//Returns a new stream. Data can not be added to a Stream manually, use a StreamController instead.
func NewStream() (s *Stream) {
	s = &Stream{
		make(chan *Subscription),
		make(chan *Subscription),
		make(chan Data),
		[]*Subscription{},
		NewCompleter(),
		make(chan struct{}),
	}
	go s.run()

	return
}

func (s *Stream) add_data(d Data) {
	for _, ss := range s.subscriptions {
		ss.add(d)
	}
}

func (s *Stream) subscribe(ss *Subscription) {
	s.subscriptions = append(s.subscriptions, ss)
}

func (s *Stream) unsubscribe(rss *Subscription) {
	index := -1
	rssp := reflect.ValueOf(rss).Pointer()
	for i, ss := range s.subscriptions {
		if reflect.ValueOf(ss).Pointer() == rssp {
			index = i
			break
		}
	}
	if index != -1 {
		l := len(s.subscriptions)
		s.subscriptions[index].close()
		switch l {
		case 1:
			s.subscriptions = []*Subscription{}
		case 2:
			if index == 0 {
				s.subscriptions[0] = s.subscriptions[1]
			}
			s.subscriptions = s.subscriptions[:1]
		default:
			s.subscriptions[index] = s.subscriptions[l-1]
			s.subscriptions = s.subscriptions[:l-1]

		}

	}
}

func (s *Stream) close() {
	s.stop <- struct{}{}
}

// Closes a Stream and cancels all subscriptions.
func (s *Stream) Close() {
	s.close()
	s.closed.Complete(nil)
}

func (s *Stream) Closed() *Future {
	return s.closed.Future()
}

// Registers a subscriber. Returns Subscription, which can be used to terminate the subcription.
func (s *Stream) Listen(sr Subscriber) (ss *Subscription) {
	ss = NewSubscription(s, sr)
	s.add_subscription <- ss
	return
}

func (s *Stream) removeSubscription(ss *Subscription) CompletionHandler {
	return func(Data) Data {
		if !s.closed.Completed() {
			s.remove_subscription <- ss
		}
		return nil
	}
}

// Registers a Transformer and returns the transformed stream.
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

// Registers a Filter and returns the filtered stream. Elements will be added if the Filter returns TRUE.
func (s *Stream) Where(f Filter) (fs *Stream) {
	fs = NewStream()
	s.Listen(filter(fs, f)).CloseOnFuture(fs.closed.Future())
	return
}

func filter(s *Stream, f Filter) Subscriber {
	return func(d Data) {
		if f(d) {
			s.add(d)
		}
	}
}

// Registers a Filter and returns the filtered stream. Elements will be added if the Filter returns FALSE.
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

// Returns a future that will be completed with the first element added to the stream.
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

// Returns a future that will be completed with the first element added to the stream where filter returns TRUE.
func (s *Stream) FirstWhere(f Filter) *Future {
	return s.Where(f).First()
}

// Returns a future that will be completed with the first element added to the stream where filter returns FALSE.
func (s *Stream) FirstWhereNot(f Filter) *Future {
	return s.WhereNot(f).First()
}

// Returns a stream with all elements where the filter returns TRUE and one where the filter returns FALSE.
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

func (s *Stream) add(d Data) {
	s.data_in <- d
	return
}

func (s *Stream) run() {
	for {

		select {

		case ss, ok := <-s.add_subscription:
			if !ok {
				return
			}
			s.subscribe(ss)

		case ss, ok := <-s.remove_subscription:
			if !ok {
				return
			}
			s.unsubscribe(ss)

		case d, ok := <-s.data_in:
			if !ok {
				return
			}
			for _, ss := range s.subscriptions {
				ss.add(d)
			}

		case <-s.stop:
			close(s.add_subscription)
			close(s.remove_subscription)
			close(s.data_in)
			for _, ss := range s.subscriptions {
				ss.close()
			}
			return
		}

	}

}
