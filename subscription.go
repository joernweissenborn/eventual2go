package eventual2go

// Subscription invokes a Subscriber when data is added to the consumed stream. It is also used for terminating a
// Subscription.
type Subscription struct {
	sr Subscriber

	inC chan Data
	doC chan Data

	closed *Completer

	unsubscribe chan *Subscription
}

// Creates a new subscription.
func NewSubscription(s *Stream,  sr Subscriber) (ss *Subscription) {
	ss = new(Subscription)
	ss.sr = sr
	ss.inC = make(chan Data)
	ss.doC = make(chan Data)
	ss.unsubscribe = s.remove_subscription
//	ss.CloseOnFuture(s.closed)
	ss.closed = NewCompleter()
	go ss.do()
	go ss.in()
	return
}

func (s *Subscription) Closed() *Future{
	return s.closed.Future()
}

func (s *Subscription) add(d Data) {
	s.inC <- d
}

func (s *Subscription) in() {
	pile := []interface{}{}
	stop := false
	for {
		if len(pile) == 0 {
			d, ok := <-s.inC

			if ok{
				pile = append(pile, d)
			} else {
				close(s.doC)
				return
			}
		} else {
			select {
			case s.doC <- pile[0]:
				pile = pile[1:]
				if len(pile) == 0 && stop{
					close(s.doC)
					return
				}
			case d, ok := <-s.inC:
				if ok {
					pile = append(pile, d)
				} else {
					stop = true
				}

			}
		}
	}
	s.closed.Complete(nil)
}

func (s *Subscription) do(){
	for d := range s.doC {
		s.sr(d)
	}
}


// Terminates the Subscription.
func (s *Subscription) Close() {
	s.unsubscribe <- s
}
func (s *Subscription) close() {
	close(s.inC)
}

// Terminates the Subscription when the Future (error-) completes.
func (s *Subscription) CloseOnFuture(f *Future) {
	f.Then(s.closeOnComplete)
	f.Err(s.closeOnCompleteError)
}

func (s *Subscription) closeOnComplete(Data)Data{
	s.Close()
	return nil
}
func (s *Subscription) closeOnCompleteError(error)(Data,error){
	s.Close()
	return nil, nil
}