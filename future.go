package eventual2go

import (
	"fmt"
	"sync"
)

// Future is thread-safe struct that can be completed with arbitrary data or failed with an error. Handler functions can
// be registered for both events and get invoked after completion..
type Future struct {
	m         *sync.Mutex
	fcs       []futurecompleter
	fces      []futurecompletererror
	completed bool
	result    Data
	err       error
}

// Creates a new future.
func newFuture() (F *Future) {
	F = &Future{
		m:         new(sync.Mutex),
		fcs:       []futurecompleter{},
		fces:      []futurecompletererror{},
		completed: false,
	}
	return
}

// Completes the future with the given data and triggers al registered completion handlers. Panics if the future is already
// complete.
func (f *Future) complete(d Data) {
	f.m.Lock()
	defer f.m.Unlock()
	if f.completed {
		panic(fmt.Sprint("Completed complete future with", d))
	}
	f.result = d
	for _, fc := range f.fcs {
		go deliverData(fc, d)
	}
	f.completed = true
}

// Completed returns the completion state.
func (f *Future) Completed() bool {
	f.m.Lock()
	defer f.m.Unlock()
	return f.completed
}

// Completes the future with the given error and triggers al registered error handlers. Panics if the future is already
// complete.
func (f *Future) completeError(err error) {
	f.m.Lock()
	defer f.m.Unlock()
	if f.completed {
		panic(fmt.Sprint("Errorcompleted complete future with", err))
	}
	f.err = err
	for _, fce := range f.fces {
		deliverErr(fce, f.err)
	}
	f.completed = true
}

func deliverData(fc futurecompleter, d Data) {
	fc.f.complete(fc.cf(d))
}

func deliverErr(fce futurecompletererror, e error) {
	go func() {
		d, err := fce.ef(e)
		if err == nil {
			fce.f.complete(d)
		} else {
			fce.f.completeError(err)
		}
	}()
}

// Then registers a completion handler. If the future is already complete, the handler gets executed immediately.
// Returns a future that gets completed with result of the handler.
func (f *Future) Then(ch CompletionHandler) (nf *Future) {
	f.m.Lock()
	defer f.m.Unlock()

	nf = newFuture()
	fc := futurecompleter{ch, nf}
	if f.completed && f.err == nil {
		deliverData(fc, f.result)
	} else if !f.completed {
		f.fcs = append(f.fcs, fc)
	}
	return
}

// WaitUntilComplete blocks until the future is complete.
func (f *Future) WaitUntilComplete() {
	<-f.AsChan()
}

// GetResult returns the result of the future, nil called before completion or after error completion.
func (f *Future) GetResult() Data {
	return f.result
}

// Err registers an error handler. If the future is already completed with an error, the handler gets executed
// immediately.
// Returns a future that either gets completed with result of the handler or error completed with the error from handler,
// if not nil.
func (f *Future) Err(eh ErrorHandler) (nf *Future) {
	f.m.Lock()
	defer f.m.Unlock()

	nf = newFuture()
	fce := futurecompletererror{eh, nf}
	if f.err != nil {
		deliverErr(fce, f.err)
	} else if !f.completed {
		f.fces = append(f.fces, fce)
	}
	return
}

// AsChan returns a channel which will receive either the result or the error after completion of the future.
func (f *Future) AsChan() chan Data {
	c := make(chan Data, 1)
	cmpl := func(d chan Data) CompletionHandler {
		return func(e Data) Data {
			d <- e
			close(d)
			return nil
		}
	}
	ecmpl := func(d chan Data) ErrorHandler {
		return func(e error) (Data, error) {
			d <- e
			close(d)
			return nil, nil
		}
	}
	f.Then(cmpl(c))
	f.Err(ecmpl(c))
	return c
}

type futurecompleter struct {
	cf CompletionHandler
	f  *Future
}

type futurecompletererror struct {
	ef ErrorHandler
	f  *Future
}
