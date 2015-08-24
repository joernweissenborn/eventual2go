package eventual2go
import (
	"testing"
	"time"
	"errors"
)

func TestReactorBasic(t *testing.T) {
	r := NewReactor()
	rt := new(reactorTester)
	r.React("TestEvent",rt.Handler)

	r.Fire("TestEvent","HALLO")

	if !rt.evtFired {
		t.Error("Event didnt fire")
	}
	if rt.data.(string) != "HALLO" {
		t.Error("Wrong Data")
	}
}

func TestReactorFuture(t *testing.T) {
	r := NewReactor()
	rt := new(reactorTester)
	r.React("TestEvent",rt.Handler)

	f := NewFuture()
	r.AddFuture("TestEvent",f)
	f.Complete("HALLO")
	time.Sleep(1*time.Millisecond)
	if !rt.evtFired {
		t.Error("Event didnt fire")
	}
	if rt.data.(string) != "HALLO" {
		t.Error("Wrong Data")
	}
}
func TestReactorFutureError(t *testing.T) {
	r := NewReactor()
	rt := new(reactorTester)
	r.React("TestEvent",rt.Handler)

	f := NewFuture()
	r.AddFutureError("TestEvent",f)
	f.CompleteError(errors.New("HALLO"))
	time.Sleep(1*time.Millisecond)
	if !rt.evtFired {
		t.Error("Event didnt fire")
	}
	if rt.data.(error).Error() != "HALLO" {
		t.Error("Wrong Data")
	}
}
func TestReactorStream(t *testing.T) {
	r := NewReactor()
	rt := new(reactorTester)
	r.React("TestEvent",rt.Handler)

	s := NewStreamController()
	r.AddStream("TestEvent",s.Stream)
	s.Add("HALLO")
	time.Sleep(1*time.Millisecond)
	if !rt.evtFired {
		t.Error("Event didnt fire")
	}
	if rt.data.(string) != "HALLO" {
		t.Error("Wrong Data")
	}
}

type reactorTester struct {
	evtFired bool
	data Data
}

func (rt *reactorTester) Handler(d Data){
	rt.evtFired = true
	rt.data = d
}