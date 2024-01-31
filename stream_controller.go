package eventual2go

import "sync"

// A StreamController is Stream where elements can be added manually or other Streams joined in.
type StreamController[T any] struct {
	m      *sync.Mutex
	stream *Stream[T]
	next   *Completer[*streamEvent[T]]
}

// NewStreamController creates a new StreamController.
func NewStreamController[T any]() (sc *StreamController[T]) {
	sc = &StreamController[T]{
		m:    &sync.Mutex{},
		next: NewCompleter[*streamEvent[T]](),
	}
	sc.stream = newStream[T](sc.next.Future())
	return
}

// Add adds an element to the stream.
func (sc *StreamController[T]) Add(d T) {
	sc.m.Lock()
	defer sc.m.Unlock()
	next := sc.getNext()
	next.Complete(&streamEvent[T]{
		data: d,
		next: sc.next.Future(),
	})
}

func (sc *StreamController[T]) getNext() (next *Completer[*streamEvent[T]]) {
	next = sc.next
	sc.next = NewCompleter[*streamEvent[T]]()
	sc.stream.updateNext(sc.next.Future())
	return
}

// Stream return the underlying stream.
func (sc *StreamController[T]) Stream() *Stream[T] {
	return sc.stream
}

// Join joins a stream. All elements from the source will be added to the stream
func (sc *StreamController[T]) Join(source *Stream[T]) {
	stop := source.Listen(sc.Add)
	stop.CompleteOnFuture(sc.stream.close.Future())
}

// JoinFuture joins a future completion event. The result will be added to the stream.
func (sc *StreamController[T]) JoinFuture(f *Future[T]) {
	f.Then(sc.Add)
}
