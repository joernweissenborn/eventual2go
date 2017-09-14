package eventual2go

import (
	"sync"
	"testing"
	"time"
)

func TestStreamBasics(t *testing.T) {
	sc := NewStreamController()
	c, _ := sc.Stream().AsChan()
	sc.Add("test")

	select {
	case <-time.After(1 * time.Millisecond):
		t.Fatal("no response")
	case data := <-c:
		if data.(string) != "test" {
			t.Error("got wrong data")
		}
	}
}

func TestStreamCancelSub(t *testing.T) {
	sc := NewStreamController()
	a := false
	stop := sc.Stream().Listen(func(d Data) {
		a = true
	})
	stop.Complete(nil)
	sc.Add(0)
	if a {
		t.Error("subscription didn't cancel")
	}

	b := false
	m := &sync.Mutex{} //protect datarace
	c := sc.Stream().Listen(func(Data) {
		b = true
	})
	c.Complete(0)
	sc.Add(0)

	m.Lock()
	defer m.Unlock()
	if b {
		t.Error("subscription didn't cancel")
	}

}

func TestStreamMultiSubscription(t *testing.T) {
	sc := NewStreamController()
	c1, _ := sc.Stream().AsChan()
	c2, _ := sc.Stream().AsChan()
	sc.Add("test")
	select {
	case <-time.After(1 * time.Millisecond):
		t.Fatal("no response")
	case data := <-c1:
		if data.(string) != "test" {
			t.Error("got wrong data")
		}
	}
	select {
	case <-time.After(1 * time.Millisecond):
		t.Fatal("no response")
	case data := <-c2:
		if data.(string) != "test" {
			t.Error("got wrong data")
		}
	}

}

func TestStreamFirst(t *testing.T) {
	sc := NewStreamController()
	c := sc.Stream().First().AsChan()

	sc.Add("test")
	select {
	case <-time.After(1 * time.Second):
		t.Error("no response")
	case data := <-c:
		if data.(string) != "test" {
			t.Error("got wrong data")
		}
	}
}

func TestStreamFilter(t *testing.T) {
	sc := NewStreamController()
	c, _ := sc.Stream().Where(func(d Data) bool { return d.(int) != 2 }).AsChan()
	sc.Add(1)
	sc.Add(2)
	sc.Add(2)
	sc.Add(1)
	sc.Add(2)
	sc.Add(5)
	for i := 0; i < 3; i++ {

		select {
		case <-time.After(1 * time.Millisecond):
			t.Fatal("no response")
		case data := <-c:
			if data.(int) == 2 {
				t.Error("got 2")
			}
		}
	}
}

func TestStreamSplit(t *testing.T) {
	sc := NewStreamController()
	y, n := sc.Stream().Split(func(d Data) bool { return d.(int) != 2 })
	c1, _ := y.AsChan()
	c2, _ := n.AsChan()
	sc.Add(1)
	sc.Add(2)
	sc.Add(2)
	sc.Add(1)
	sc.Add(2)
	sc.Add(5)
	for i := 0; i < 3; i++ {
		select {
		case <-time.After(1 * time.Millisecond):
			t.Fatal("no response")
		case data := <-c1:
			if data.(int) == 2 {
				t.Error("got 2")
			}
		}
	}
	for i := 0; i < 3; i++ {
		select {
		case <-time.After(1 * time.Millisecond):
			t.Fatal("no response")
		case data := <-c2:
			if data.(int) != 2 {
				t.Error("got 2")
			}
		}
	}
}

func TestStreamTransformer(t *testing.T) {
	sc := NewStreamController()
	c, _ := sc.Stream().Transform(func(d Data) Data { return d.(int) * 2 }).AsChan()
	sc.Add(5)
	select {
	case <-time.After(1 * time.Millisecond):
		t.Fatal("no response")
	case data := <-c:
		if data.(int) != 10 {
			t.Error("got 5")
		}
	}
}

func TestJoinFuture(t *testing.T) {

	sc := NewStreamController()
	c, _ := sc.Stream().AsChan()
	fc := NewCompleter()
	sc.JoinFuture(fc.Future())
	fc.Complete("test")
	select {
	case <-time.After(1 * time.Second):
		t.Error("no response")
	case data := <-c:
		if data.(string) != "test" {
			t.Error("got wrong data")
		}
	}
}
