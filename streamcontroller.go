package eventual2go

import "fmt"

// A StreamController is Stream where elements can be added manually or other Streams joined in.
type StreamController struct {
	s *Stream
}

// Add adds an element to the stream.
func (sc *StreamController) Add(Data Data) {
	if sc.s.closed == nil {
		panic("Add on noninitialized StreamController")
	}
	if sc.s.closed.Completed() {
		panic(fmt.Sprint("Add on closed stream", Data))
	}
	sc.s.add(Data)
}

// Stream return the underlying stream.
func (sc *StreamController) Stream() *Stream {
	return sc.s
}

// Close closes the underlying stream.
func (sc *StreamController) Close() {
	sc.s.Close()
}

// Closed returns a future which is completed when the underlying stream is closed.
func (sc *StreamController) Closed() *Future {
	return sc.s.Closed()
}

// NewStreamController creates a new StreamController.
func NewStreamController() (sc *StreamController) {
	sc = &StreamController{NewStream()}
	if sc.s.Closed() == nil {
		panic("Stream Init failed")
	}
	return
}

// Join joins a stream. All elements from the source will be added to the stream
func (sc *StreamController) Join(s *Stream) {
	if s.closed == nil {
		panic("Join noninitialized Stream")
	}
	if s.closed.Completed() {
		panic("Join closed Stream")
	}
	if sc.s.closed == nil {
		panic("Join on noninitialized Streamcontroller")
	}
	if sc.s.closed.Completed() {
		panic("Join on closed Streamcontroller")
	}
	ss := s.Listen(addJoined(sc))
	ss.CloseOnFuture(sc.s.closed.Future())
}

// JoinFuture joins a future completion event.
func (sc *StreamController) JoinFuture(f *Future) {
	if sc.s.closed == nil {
		panic("Join on noninitialized Streamcontroller")
	}
	if sc.s.closed.Completed() {
		panic("Join on closed Streamcontroller")
	}
	f.Then(addJoinedFuture(sc))
}

func addJoined(sc *StreamController) Subscriber {
	return func(d Data) {
		sc.s.add(d)
	}
}

func addJoinedFuture(sc *StreamController) CompletionHandler {
	return func(d Data) Data {
		sc.s.add(d)
		return nil
	}
}
