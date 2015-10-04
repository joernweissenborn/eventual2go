package eventual2go

import (
	"fmt"
	"sync"
)

// Creates a new future.
func NewFuture() (F *Future) {
	F = &Future{
		new(sync.Mutex),
		[]futurecompleter{},
		[]futurecompletererror{},
		false,
		nil,
		nil,
	}
	return
}

// Future is thread-safe struct that can be completed with arbitrary data or failed with an error. Handler functions can
// be registered for both events and get invoked after completion..
type Future struct {
	m         *sync.Mutex
	fcs       []futurecompleter
	fces      []futurecompletererror
	completed bool
	r         interface{}
	e         error
}

// Completes the future with the given data and triggers al registered completion handlers. Panics if the future is already
// complete.
func (f *Future) complete(d Data) {
	f.m.Lock()
	defer f.m.Unlock()
	if f.completed {
		panic(fmt.Sprint("Completed complete future with", d))
	}
	f.r = d
	for _, fc := range f.fcs {
		go deliverData(fc, d)
	}
	f.completed = true
}

// Returns the completion state.
func (f *Future) Completed() bool {
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
	f.e = err
	for _, fce := range f.fces {
		deliverErr(fce, f.e)
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

	nf = NewFuture()
	fc := futurecompleter{ch, nf}
	if f.completed && f.e == nil {
		deliverData(fc, f.r)
	} else if !f.completed {
		f.fcs = append(f.fcs, fc)
	}
	return
}

// Blocks until the future is complete.
func (f *Future) WaitUntilComplete() {
	<-f.AsChan()
}

// Returns the result of the future, nil called before completion or after error completion.
func (f *Future) GetResult() Data {
	return f.r
}

// Then registers an error handler. If the future is already completed with an error, the handler gets executed
// immediately.
// Returns a future that either gets completed with result of the handler or error completed with the error from handler,
// if not nil.
func (f *Future) Err(eh ErrorHandler) (nf *Future) {
	f.m.Lock()
	defer f.m.Unlock()

	nf = NewFuture()
	fce := futurecompletererror{eh, nf}
	if f.e != nil {
		deliverErr(fce, f.e)
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
