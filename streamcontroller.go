package eventual2go

// A StreamController is Stream where elements can be added manually or other Streams joined in.
type StreamController struct {
	*Stream
}

// Adds an element to the stream.
func (sc *StreamController) Add(Data Data) {
	if sc.closed == nil {
		panic("Add on noninitialized StreamController")
	}
	if sc.Stream.closed.Completed() {
		panic("Add on closed stream")
	}
	sc.Stream.add(Data)
}

// Creates a new StreamController.
func NewStreamController() (sc *StreamController) {
	sc = &StreamController{sc.Stream}
	if sc.Stream.Closed() == nil {
		panic("Stream Init failed")
	}
	return
}

// Joins a stream. All elements from the source will be added to the stream
func (sc *StreamController) Join(s *Stream) {
	if s.closed == nil {
		panic("Join noninitialized Stream")
	}
	if s.closed.Completed() {
		panic("Join closed Stream")
	}
	if sc.closed == nil {
		panic("Join on noninitialized Streamcontroller")
	}
	if sc.closed.Completed() {
		panic("Join on closed Streamcontroller")
	}
	ss := s.Listen(addJoined(sc))
	ss.CloseOnFuture(sc.closed.Future())
}

// Joins a future completion event.
func (sc *StreamController) JoinFuture(f *Future) {
	if sc.closed == nil {
		panic("Join on noninitialized Streamcontroller")
	}
	if sc.closed.Completed() {
		panic("Join on closed Streamcontroller")
	}
	f.Then(addJoinedFuture(sc))
}

func addJoined(sc *StreamController) Subscriber {
	return func(d Data) {
		sc.Stream.add(d)
	}
}

func addJoinedFuture(sc *StreamController) CompletionHandler {
	return func(d Data) Data {
		sc.Stream.add(d)
		return nil
	}
}
