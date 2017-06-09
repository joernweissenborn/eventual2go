package eventual2go

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestReactorBasic(t *testing.T) {
	r := NewReactor()
	rt := &reactorTester{Mutex: &sync.Mutex{}}
	r.React("TestEvent", rt.Handler)

	r.Fire("TestEvent", "HALLO")
	time.Sleep(1 * time.Millisecond)

	if !rt.hasFired() {
		t.Fatal("Event didnt fire")
	}
	if rt.data.(string) != "HALLO" {
		t.Error("Wrong Data")
	}
	r.Shutdown(nil)
	if !r.ShutdownFuture().WaitUntilTimeout(1 * time.Millisecond) {
		t.Error("Rector shutdown failed")
	}
}

func TestReactorMultipleEvents(t *testing.T) {
	r := NewReactor()
	rt1 := &reactorTester{Mutex: &sync.Mutex{}}
	rt2 := &reactorTester{Mutex: &sync.Mutex{}}
	r.React("TestEvent1", rt1.Handler)
	r.React("TestEvent2", rt2.Handler)
	r.Fire("TestEvent2", "HALLO")
	r.Fire("TestEvent1", "HALLO")
	time.Sleep(1 * time.Millisecond)

	if !rt1.hasFired() {
		t.Fatal("Event didnt fire")
	}
	if rt1.data.(string) != "HALLO" {
		t.Error("Wrong Data")
	}

	if !rt2.evtFired {
		t.Fatal("Event didnt fire")
	}
	if rt2.data.(string) != "HALLO" {
		t.Error("Wrong Data")
	}
}

func TestReactorFuture(t *testing.T) {
	r := NewReactor()
	rt := &reactorTester{Mutex: &sync.Mutex{}}

	r.React("TestEvent", rt.Handler)

	c := NewCompleter()
	f := c.Future()
	r.AddFuture("TestEvent", f)
	c.Complete("HALLO")
	time.Sleep(1 * time.Millisecond)
	if !rt.hasFired() {
		t.Fatal("Event didnt fire")
	}
	if rt.data.(string) != "HALLO" {
		t.Error("Wrong Data")
	}
}
func TestReactorFutureError(t *testing.T) {
	r := NewReactor()
	rt := &reactorTester{Mutex: &sync.Mutex{}}

	r.React("TestEvent", rt.Handler)

	c := NewCompleter()
	f := c.Future()
	r.AddFutureError("TestEvent", f)
	c.CompleteError(errors.New("HALLO"))
	time.Sleep(1 * time.Millisecond)
	if !rt.hasFired() {
		t.Fatal("Event didnt fire")
	}
	if rt.data.(error).Error() != "HALLO" {
		t.Error("Wrong Data")
	}
}
func TestReactorStream(t *testing.T) {
	r := NewReactor()
	rt := &reactorTester{Mutex: &sync.Mutex{}}
	r.React("TestEvent", rt.Handler)

	s := NewStreamController()
	r.AddStream("TestEvent", s.Stream())
	s.Add("HALLO")
	time.Sleep(1 * time.Millisecond)
	if !rt.hasFired() {
		t.Fatal("Event didnt fire")
	}
	if rt.data.(string) != "HALLO" {
		t.Error("Wrong Data")
	}
}

type reactorTester struct {
	*sync.Mutex
	evtFired bool
	data     Data
}

func (rt *reactorTester) Handler(d Data) {
	rt.Lock()
	defer rt.Unlock()
	rt.evtFired = true
	rt.data = d
}

func (rt *reactorTester) hasFired() bool {
	rt.Lock()
	defer rt.Unlock()
	return rt.evtFired
}
