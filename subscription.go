package eventual2go

import "sync"

// Subscription invokes a Subscriber when data is added to the consumed stream. It is also used for terminating a
// Subscription.
type Subscription struct {
	canceled *Completer
	closed   *Completer
	sr       Subscriber
	inC      chan Data
	doC      chan Data
	m        *sync.Mutex

	next Future
}

// NewSubscription creates a new subscription.
func NewSubscription(s *Stream, sr Subscriber) (ss *Subscription) {
	ss = &Subscription{
		canceled: NewCompleter(),
		closed:   NewCompleter(),
		sr:       sr,
		inC:      make(chan Data),
		doC:      make(chan Data),
		m:        &sync.Mutex{},
	}
	go ss.do()
	go ss.in()
	return
}

// Closed returns a future which gets completed when subscription is closed.
func (s *Subscription) Closed() *Future {
	return s.closed.Future()
}

func (s *Subscription) add(d Data) {
	s.m.Lock()
	defer s.m.Unlock()
	if s.canceled.Completed() {
		return
	}
	s.inC <- d
}

func (s *Subscription) in() {
	pile := []interface{}{}
	stop := false
	for {
		if len(pile) == 0 {
			d, ok := <-s.inC

			if ok && !stop {
				pile = append(pile, d)
			} else {
				close(s.doC)
				return
			}
		} else {
			select {
			case s.doC <- pile[0]:
				if len(pile) == 1 && stop {
					close(s.doC)
					return
				}
				pile = pile[1:]
			case d, ok := <-s.inC:
				if ok {
					pile = append(pile, d)
				} else {
					stop = true
				}

			}
		}
	}
}

func (s *Subscription) do() {
	for d := range s.doC {
		s.sr(d)
	}
	s.closed.Complete(s)
}

// Close terminates the Subscription.
func (s *Subscription) Close() {
	s.m.Lock()
	defer s.m.Unlock()
	if s.canceled.Completed() {
		return
	}
	s.canceled.Complete(s)
	close(s.inC)
}

// CloseOnFuture terminates the Subscription when the given Future (error-)completes.
func (s *Subscription) CloseOnFuture(f *Future) {
	f.Then(s.closeOnComplete)
	f.Err(s.closeOnCompleteError)
}

func (s *Subscription) closeOnComplete(Data) Data {
	s.Close()
	return nil
}
func (s *Subscription) closeOnCompleteError(error) (Data, error) {
	s.Close()
	return nil, nil
}
