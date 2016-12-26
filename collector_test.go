package eventual2go

import (
	"testing"
	"time"
)

func TestCollector(t *testing.T) {

	c := NewCollector()

	if !c.Empty() {
		t.Error("Collector is not empty after init")
	}

	c.Add("bla")

	//Add is non-blocking async
	time.Sleep(1 * time.Millisecond)

	if c.Empty() {
		t.Error("Collector is still empty after add")
	}

	d, ok := c.Preview().(string)

	if !ok {
		t.Error("Wrong Data Type", d)
	}
	if d != "bla" {
		t.Error("Wrong data", d)
	}

	if c.Empty() {
		t.Error("Collector is empty after preview")
	}

	d, ok = c.Get().(string)

	if !ok {
		t.Error("Wrong Data Type", d)
	}
	if d != "bla" {
		t.Error("Wrong data", d)
	}

	if !c.Empty() {
		t.Error("Collector is not empty after get")
	}

	f := NewCompleter()
	c.AddFuture(f.Future())
	f.Complete("bla")

	//Add is non-blocking async
	time.Sleep(1 * time.Millisecond)

	if c.Empty() {
		t.Error("Collector is still empty after add")
	}

	d, ok = c.Get().(string)

	if !ok {
		t.Error("Wrong Data Type", d)
	}
	if d != "bla" {
		t.Error("Wrong data", d)
	}
}
