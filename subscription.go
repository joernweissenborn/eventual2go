package eventual2go

// Creates a new subscription.
func NewSubscription(index int, close chan int, sr Subscriber) (s Subscription) {
	s.in = make(chan Data)
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
	in    chan Data
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

// Terminates the Subscription when the Future (error-) completes.
func (s Subscription) CloseOnFuture(f *Future) {
	f.Then(s.closeOnComplete)
	f.Err(s.closeOnCompleteError)
}

func (s Subscription) closeOnComplete(Data)Data{
	s.Close()
	return nil
}
func (s Subscription) closeOnCompleteError(error)(Data,error){
	s.Close()
	return nil, nil
}