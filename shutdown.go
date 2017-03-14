package eventual2go

import "sync"

type Shutdowner interface {
	Shutdown(d Data) error
}

type Shutdown struct {
	m      *sync.Mutex
	errors []error
	initSd *Completer
	wg     *sync.WaitGroup
}

func NewShutdown() (sd *Shutdown) {
	sd = &Shutdown{
		m:      &sync.Mutex{},
		errors: []error{},
		initSd: NewCompleter(),
		wg:     &sync.WaitGroup{},
	}
	return
}

func (sd *Shutdown) Register(s Shutdowner) {
	sd.initSd.Future().Then(sd.doShutdown(s))
	sd.wg.Add(1)
}

func (sd *Shutdown) Do(d Data) (errs []error) {
	sd.initSd.Complete(d)
	sd.wg.Wait()
	errs = sd.errors
	return
}

func (sd *Shutdown) doShutdown(s Shutdowner) CompletionHandler {
	return func(d Data) Data {
		if err := s.Shutdown(d); err != nil {
			sd.m.Lock()
			sd.errors = append(sd.errors, err)
			sd.m.Unlock()
		}
		sd.wg.Done()
		return nil
	}
}
