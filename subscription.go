package eventual2go

// Creates a new subscription.
func NewSubscription(index int, close chan int, sr Subscriber) (s Subscription) {
	s.in = make(chan interface{})
	s.add = make(streamchannel)
	go s.add.pipe(s.in)
	s.index = index
	s.close = close
	s.sr = sr
	go s.run()
	return
}

// Subscription invokes a Subscriber when data is added to the consumed stream. It is also used for terminating a
// Subscription.
type Subscription struct {
	in    chan interface{}
	add   streamchannel
	index int
	close chan int

	sr Subscriber
}

func (s Subscription) run() {
	for d := range s.in {
		s.sr(d)
	}
}

// Terminates the Subscription.
func (s Subscription) Close() {
	defer func() { recover() }()
	s.close <- s.index

}