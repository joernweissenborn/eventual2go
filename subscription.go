package eventual2go

// Subscription invokes a Subscriber when data is added to the consumed stream. It is also used for terminating a
// Subscription.
type Subscription struct {
	sr Subscriber

	inC chan Data
	doC chan Data

	closed *Future

}

// Creates a new subscription.
func NewSubscription(s *Stream,  sr Subscriber) (ss *Subscription) {
	ss = new(Subscription)
	ss.sr = sr
	ss.inC = make(chan Data)
	ss.doC = make(chan Data)
//	ss.CloseOnFuture(s.closed)
	ss.closed = NewFuture()
	go ss.do()
	go ss.in()
	return
}

func (s *Subscription) add(d Data) {
	s.inC <- d
}

func (s *Subscription) in() {
	pile := []interface{}{}
	for {
		if len(pile) == 0 {
			pile = append(pile, <-s.inC)
		} else {
			select {
			case s.doC <- pile[0]:
				pile = pile[1:]
				if len(pile) == 0 && s.closed.IsComplete(){
					close(s.doC)
					return
				}
			case d, ok := <-s.inC:
				if ok {
					pile = append(pile, d)
				}

			}
		}
	}
}

func (s *Subscription) do(){
	for d := range s.doC {
		s.sr(d)
	}
	s.closed.Complete(nil)
}


// Terminates the Subscription.
func (s *Subscription) Close() {
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