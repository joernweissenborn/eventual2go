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

func (fwg *FutureWaitGroup) Add(f *Future) {
	fwg.wg.Add(1)
	f.Then(fwg.onComplete)
	f.Err(fwg.onErrComplete)
}

func (fwg *FutureWaitGroup) onComplete(Data) Data {
	fwg.wg.Done()
	return nil
}

func (fwg *FutureWaitGroup) onErrComplete(error) (Data, error) {
	fwg.wg.Done()
	return nil, nil
}
