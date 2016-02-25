package eventual2go

type Shutdowner interface {
	Shutdown(d Data) error
}

type Shutdown struct {
	errCollector *Collector
	initSd       *Completer
	wg           *FutureWaitGroup
}

func NewShutdown() (sd *Shutdown) {
	sd = &Shutdown{
		errCollector: NewCollector(),
		initSd:       NewCompleter(),
		wg:           NewFutureWaitGroup(),
	}
	return
}

func (sd *Shutdown) Register(s Shutdowner) {
	c := NewCompleter()

	sd.initSd.Future().Then(doShutdown(c, s))
	sd.wg.Add(c.Future())
	sd.errCollector.AddFutureError(c.Future())
}

func (sd *Shutdown) Do(d Data) (errs []error) {

	sd.initSd.Complete(d)

	sd.wg.Wait()

	for !sd.errCollector.Empty() {
		errs = append(errs, sd.errCollector.Get().(error))
	}
	return
}

func doShutdown(c *Completer, s Shutdowner) CompletionHandler {
	return func(d Data) Data {
		if err := s.Shutdown(d); err != nil {
			c.CompleteError(err)
		} else {
			c.Complete(nil)
		}
		return nil
	}
}
