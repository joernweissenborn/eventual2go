package eventual2go

// A StreamController is Stream where elements can be added manually or other Streams joined in.
type StreamController struct {
	Stream
}

// Adds an element to the stream.
func (sc StreamController) Add(Data Data) {
	if sc.Closed == nil {
		panic("Add on noninitialized StreamController")
	}
	if sc.Stream.Closed.IsComplete() {
		panic("Add on closed stream")
	}
	sc.Stream.add(Data)
}

// Creates a new StreamController.
func NewStreamController() (sc StreamController) {
	sc.Stream = NewStream()
	if sc.Stream.Closed == nil {
		panic("Stream Init failed")
	}
	return
}

// Joins a stream. All elements from the source will be added to the stream
func (sc StreamController) Join(s Stream) {
	if s.Closed == nil {
		panic("Join noninitialized Stream")
	}
	if sc.Closed == nil {
		panic("Join on noninitialized Streamcontroller")
	}
	ss := s.Listen(addJoined(sc))
	s.Closed.Then(closeSus(ss))
}

func closeSus(ss Subscription) CompletionHandler {
	return func(Data) Data {
		ss.Close()
		return nil
	}
}

func addJoined(sc StreamController) Subscriber {
	return func(d Data) {
		sc.Stream.add(d)
	}
}
