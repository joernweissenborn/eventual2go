package eventual2go

// A StreamController is Stream where elements can be added manually or other Streams joined in.
type StreamController struct {
	stream *Stream
	next   *Completer
}

// NewStreamController creates a new StreamController.
func NewStreamController() (sc *StreamController) {
	sc = &StreamController{
		next: NewCompleter(),
	}
	sc.stream = newStream(sc.next.Future())
	return
}

// Add adds an element to the stream.
func (sc *StreamController) Add(d Data) {
	next := NewCompleter()
	sc.next.Complete(&streamEvent{
		data: d,
		next: next.Future(),
	})
	sc.stream.updateNext(next.Future())
	sc.next = next
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

// JoinFuture joins a future completion event.
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
