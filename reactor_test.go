package eventual2go

import (
	"errors"
	"testing"
	"time"
)

func TestReactorBasic(t *testing.T) {
	r := NewReactor()
	rt := new(reactorTester)
	r.React("TestEvent", rt.Handler)

	r.Fire("TestEvent", "HALLO")
	time.Sleep(1 * time.Millisecond)

	if !rt.evtFired {
		t.Fatal("Event didnt fire")
	}
	if rt.data.(string) != "HALLO" {
		t.Error("Wrong Data")
	}
}

func TestReactorMultipleEvents(t *testing.T) {
	r := NewReactor()
	rt1 := new(reactorTester)
	rt2 := new(reactorTester)
	r.React("TestEvent1", rt1.Handler)
	r.React("TestEvent2", rt2.Handler)
	r.Fire("TestEvent2", "HALLO")
	r.Fire("TestEvent1", "HALLO")
	time.Sleep(1 * time.Millisecond)

	if !rt1.evtFired {
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
	rt := new(reactorTester)
	r.React("TestEvent", rt.Handler)

	c := NewCompleter()
	f := c.Future()
	r.AddFuture("TestEvent", f)
	c.Complete("HALLO")
	time.Sleep(1 * time.Millisecond)
	if !rt.evtFired {
		t.Fatal("Event didnt fire")
	}
	if rt.data.(string) != "HALLO" {
		t.Error("Wrong Data")
	}
}
func TestReactorFutureError(t *testing.T) {
	r := NewReactor()
	rt := new(reactorTester)
	r.React("TestEvent", rt.Handler)

	c := NewCompleter()
	f := c.Future()
	r.AddFutureError("TestEvent", f)
	c.CompleteError(errors.New("HALLO"))
	time.Sleep(1 * time.Millisecond)
	if !rt.evtFired {
		t.Fatal("Event didnt fire")
	}
	if rt.data.(error).Error() != "HALLO" {
		t.Error("Wrong Data")
	}
}
func TestReactorStream(t *testing.T) {
	r := NewReactor()
	rt := new(reactorTester)
	r.React("TestEvent", rt.Handler)

	s := NewStreamController()
	r.AddStream("TestEvent", s.Stream())
	s.Add("HALLO")
	time.Sleep(1 * time.Millisecond)
	if !rt.evtFired {
		t.Fatal("Event didnt fire")
	}
	if rt.data.(string) != "HALLO" {
		t.Error("Wrong Data")
	}
}

type reactorTester struct {
	evtFired bool
	data     Data
}

func (rt *reactorTester) Handler(d Data) {
	rt.evtFired = true
	rt.data = d
}
