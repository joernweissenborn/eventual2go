package eventual2go

import "sync"

// Shutdowner represents the eventual2go shutdown interface.
type Shutdowner interface {
	Shutdown(d Data) error
}

// Shutdown is a register for `Shutdowner` and is used to orchestrate a concurrent shutdown.
type Shutdown struct {
	m      *sync.Mutex
	errors []error
	initSd *Completer
	wg     *sync.WaitGroup
}

// NewShutdown creats a new `Shutdown`.
func NewShutdown() (sd *Shutdown) {
	sd = &Shutdown{
		m:      &sync.Mutex{},
		errors: []error{},
		initSd: NewCompleter(),
		wg:     &sync.WaitGroup{},
	}
	return
}

// Register registers a `Shutdowner`.
func (sd *Shutdown) Register(s Shutdowner) {
	sd.initSd.Future().Then(sd.doShutdown(s))
	sd.wg.Add(1)
}

// Do intiatates the shutdown by concurrently calling the shutdown method on all registered `Shutdowner`. Blocks until all shutdowns have finished.
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
