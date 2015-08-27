package eventual2go

import (
	"errors"
	"testing"
)

func TestFutureBasicCompletion(t *testing.T) {
	cp := NewCompleter()
	f := cp.Future()

	if f.Completed() {
		t.Error("Complete to early")
	}

	c := f.AsChan()

	cp.Complete(true)

	//
	if !f.Completed() {
		t.Error("not completed")
	}

	if !(<-c).(bool) {
		t.Error("Completed with wrong args")
	}
}

func TestFutureChainCompletion(t *testing.T) {
	cp := NewCompleter()
	f := cp.Future()

	c := f.AsChan()

	cp.Complete(true)

	//
	if !f.Completed() {
		t.Error("not completed")
	}

	if !(<-c).(bool) {
		t.Error("Completed with wrong args")
	}
}

func TestFutureErrCompletion(t *testing.T) {
	cp := NewCompleter()
	f := cp.Future()

	c1 := make(chan error)
	defer close(c1)
	f.Err(testcompletererr(c1))

	cp.CompleteError(errors.New("testerror"))

	//
	if !f.Completed() {
		t.Error("not completed")
	}

	c2 := make(chan error)
	defer close(c2)
	c3 := make(chan interface{})
	defer close(c3)
	f.Err(testcompletererr(c2)).Then(testcompleter(c3))

	if (<-c1).Error() != "testerror" {
		t.Error("Completed with wrong err")
	}

	if (<-c2).Error() != "testerror" {
		t.Error("Completed with wrong err")
	}
	if !(<-c3).(bool) {
		t.Error("Completed with wrong args")
	}

}

func TestFutureMultiCompletion(t *testing.T) {
	cp := NewCompleter()
	f := cp.Future()

	c1 := make(chan interface{})
	defer close(c1)
	f.Then(testcompleter(c1))
	c2 := make(chan interface{})
	defer close(c2)
	f.Then(testcompleter(c2))

	if len(f.fcs) != 2 {
		t.Fatal("fcs lenth not 2, is", len(f.fcs))
	}

	cp.Complete(true)

	if !(<-c1).(bool) {
		t.Error("Completed with wrong args")
	}
	if !(<-c2).(bool) {
		t.Error("Completed with wrong args")
	}
}

func testcompleter(c chan interface{}) CompletionHandler {
	return func(d Data) Data {
		c <- d
		return nil
	}
}

func testcompletererr(c chan error) ErrorHandler {
	return func(e error) (Data, error) {
		c <- e
		return true, nil
	}
}

func testcompletererrwitherr(c chan error) ErrorHandler {
	return func(e error) (Data, error) {
		c <- e
		return true, errors.New("followerr")
	}
}
