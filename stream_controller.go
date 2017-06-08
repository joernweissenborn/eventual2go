package eventual2go

import "sync"

// A StreamController is Stream where elements can be added manually or other Streams joined in.
type StreamController struct {
	m      *sync.Mutex
	stream *Stream
	next   *Completer
}

// NewStreamController creates a new StreamController.
func NewStreamController() (sc *StreamController) {
	sc = &StreamController{
		m:    &sync.Mutex{},
		next: NewCompleter(),
	}
	sc.stream = newStream(sc.next.Future())
	return
}

// Add adds an element to the stream.
func (sc *StreamController) Add(d Data) {
	sc.m.Lock()
	defer sc.m.Unlock()
	next := sc.getNext()
	next.Complete(&streamEvent{
		data: d,
		next: sc.next.Future(),
	})
}

func (sc *StreamController) getNext() (next *Completer) {
	next = sc.next
	sc.next = NewCompleter()
	sc.stream.updateNext(sc.next.Future())
	return
}

// Stream return the underlying stream.
func (sc *StreamController) Stream() *Stream {
	return sc.stream
}

// Join joins a stream. All elements from the source will be added to the stream
func (sc *StreamController) Join(source *Stream) {
	stop := source.Listen(addJoined(sc))
	stop.CompleteOnFuture(sc.stream.close.Future())
}

// JoinFuture joins a future completion event. The result will be added to the stream.
func (sc *StreamController) JoinFuture(f *Future) {
	f.Then(addJoinedFuture(sc))
}

func addJoined(sc *StreamController) Subscriber {
	return func(d Data) {
		sc.Add(d)
	}
}

func addJoinedFuture(sc *StreamController) CompletionHandler {
	return func(d Data) Data {
		sc.Add(d)
		return nil
	}
}
