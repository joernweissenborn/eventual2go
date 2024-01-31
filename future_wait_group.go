package eventual2go

import "sync"

type FutureWaitGroup struct {
	wg sync.WaitGroup
}

func NewFutureWaitGroup() (fwg *FutureWaitGroup) {
	fwg = &FutureWaitGroup{}
	return
}

func (fwg *FutureWaitGroup) Wait() {
	fwg.wg.Wait()
}

func (fwg *FutureWaitGroup) Add(f *Future[Data]) {
	fwg.wg.Add(1)
	f.Then(fwg.onComplete)
	f.Err(fwg.onErrComplete)
}

func (fwg *FutureWaitGroup) onComplete(Data) {
	fwg.wg.Done()
}

func (fwg *FutureWaitGroup) onErrComplete(error) {
	fwg.wg.Done()
}
