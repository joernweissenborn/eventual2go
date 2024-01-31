package eventual2go

import (
	"errors"
	"testing"
	"time"
)

func TestFutureBasicCompletion(t *testing.T) {
	cp := NewCompleter[bool]()
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

	if !(<-c) {
		t.Error("Completed with wrong args")
	}

	if !f.Result(){
		t.Error("Completed with wrong args")
	}
}

func TestTimeoutCompletion(t *testing.T) {
	cp := NewTimeoutCompleter[bool](1 * time.Microsecond)
	f := cp.Future()
	time.Sleep(10 * time.Millisecond)
	if !f.Completed() {
		t.Error("Timeout didnt complete")
	}
	c := make(chan error, 1)
	defer close(c)
	f.Err(testcompletererr(c))

	if <-c != ErrTimeout {
		t.Error("Completed with wrong error")
	}
}

func TestFutureChainCompletion(t *testing.T) {
	cp := NewCompleter[bool]()
	f := cp.Future()

	c := f.AsChan()

	cp.Complete(true)

	if !f.Completed() {
		t.Error("not completed")
	}

	if !(<-c) {
		t.Error("Completed with wrong args")
	}
}

func TestFutureErrCompletion(t *testing.T) {
	cp := NewCompleter[bool]()
	f := cp.Future()

	c1 := make(chan error, 1)
	defer close(c1)
	f.Err(testcompletererr(c1))

	cp.CompleteError(errors.New("testerror"))

	if !f.Completed() {
		t.Error("not completed")
	}

	c2 := make(chan error, 1)
	defer close(c2)
	f.Err(testcompletererr(c2))

	if (<-c1).Error() != "testerror" {
		t.Error("Completed with wrong err")
	}

	if (<-c2).Error() != "testerror" {
		t.Error("Completed with wrong err")
	}

}

func TestFutureMultiCompletion(t *testing.T) {
	cp := NewCompleter[Data]()
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

func testcompleter(c chan interface{}) CompletionHandler[Data] {
	return func(d Data) {
		c <- d
	}
}

func testcompletererr(c chan error) ErrorHandler {
	return func(e error)  {
		c <- e
	}
}
