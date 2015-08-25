package eventual2go

//Returns a new stream. Data can not be added to a Stream manually, use a StreamController instead.
func NewStream() (s Stream) {
	s.in = make(chan interface{})
	s.addsus = make(chan addsuscrption)
	s.rmsusc = make(chan int)
	s.Closed = NewFuture()
	go s.run()
	return
}

// A Stream can be consumed or new streams be derived by registering handler functions.
type Stream struct {
	in chan interface{}

	addsus chan addsuscrption
	rmsusc chan int

	Closed *Future
}

type addsuscrption struct {
	sr Subscriber
	c  chan Subscription
}

// Closes a Stream and cancels all subscriptions.
func (s Stream) Close() {
	close(s.addsus)
	close(s.rmsusc)
	close(s.in)
	s.Closed.Complete(nil)
}

func (s Stream) run() {

	suscribtions := map[int]Subscription{}
	nextSusIndex := 0

	ok := true

	for ok {
		select {

		case sus, ok := <-s.addsus:
			if !ok {
				return
			}
			sc := NewSubscription(nextSusIndex, s.rmsusc, sus.sr)
			suscribtions[nextSusIndex] = sc
			sus.c <- sc
			nextSusIndex++

		case index, ok := <-s.rmsusc:
			if !ok {
				return
			}
			delete(suscribtions, index)

		case d, ok := <-s.in:
			if !ok {
				return
			}
			for _, sc := range suscribtions {
				sc.add <- d
			}
		}
	}
}

func (s Stream) add(d interface{}) {
	s.in <- d
	return
}

// Registers a subscriber. Returns Subscription, which can be used to terminate the subcription.
func (s Stream) Listen(sr Subscriber) (ss Subscription) {
	if s.Closed == nil {
		panic("Listen on uninitialized stream")
	}
	c := make(chan Subscription)
	s.addsus <- addsuscrption{sr, c}
	return <-c
}

// Registers a Transformer and returns the transformed stream.
func (s Stream) Transform(t Transformer) (ts Stream) {
	ts = NewStream()
	s.Listen(func(d Data) {
		ts.add(t(d))
	})
	return
}

// Registers a Filter and returns the filtered stream. Elements will be added if the Filter returns TRUE.
func (s Stream) Where(f Filter) (fs Stream) {
	fs = NewStream()
	s.Listen(func(d Data) {
		if f(d) {
			fs.add(d)
		}
	})

	return
}

// Registers a Filter and returns the filtered stream. Elements will be added if the Filter returns FALSE.
func (s Stream) WhereNot(f Filter) (fs Stream) {
	fs = NewStream()
	s.Listen(func(d Data) {
		if !f(d) {
			fs.add(d)
		}
	})

	return
}

// Returns a future that will be completed with the first element added to the stream.
func (s Stream) First() (f *Future) {
	f = NewFuture()
	ss := s.Listen(func(d Data) {
		if !f.IsComplete(){
			f.Complete(d)
		}
	})
	f.Then(removeSub(ss))
	return
}

func removeSub(ss Subscription) CompletionHandler{
	return func(Data) Data{
		ss.Close()
		return nil
	}
}

// Returns a future that will be completed with the first element added to the stream where filter returns TRUE.
func (s Stream) FirstWhere(f Filter) *Future {
	return s.Where(f).First()
}

// Returns a future that will be completed with the first element added to the stream where filter returns FALSE.
func (s Stream) FirstWhereNot(f Filter) *Future {
	return s.WhereNot(f).First()
}

// Returns a stream with all elements where the filter returns TRUE and one where the filter returns FALSE.
func (s Stream) Split(f Filter) (ts Stream, fs Stream) {
	return s.Where(f), s.WhereNot(f)
}

// AsChan returns a channel where all items will be pushed. Note items while be queued in a fifo since the stream must
// not block.
func (s Stream) AsChan() (c chan Data){
	c = make(chan Data)
	s.Closed.Then(closeChan(c))
	s.Listen(pipeToChan(c))
	return
}

func pipeToChan(c chan Data) Subscriber{
	return func(d Data){
		c<-d
	}
}
func closeChan(c chan Data) CompletionHandler{
	return func(d Data)Data{
		close(c)
		return nil
	}
}