package eventual2go

import (
	"sync"
	"testing"
	"time"
)

func TestStreamBasics(t *testing.T) {
	sc := NewStreamController[string]()
	c, _ := sc.Stream().AsChan()
	sc.Add("test")

	select {
	case <-time.After(1 * time.Millisecond):
		t.Fatal("no response")
	case data := <-c:
		if data != "test" {
			t.Error("got wrong data")
		}
	}
}

func TestStreamCancelSub(t *testing.T) {
	sc := NewStreamController[int]()
	a := false
	stop := sc.Stream().Listen(func(d int) {
		a = true
	})
	stop.Complete(nil)
	sc.Add(0)
	if a {
		t.Error("subscription didn't cancel")
	}

	b := false
	m := &sync.Mutex{} //protect datarace
	c := sc.Stream().Listen(func(int) {
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
	sc := NewStreamController[string]()
	c1, _ := sc.Stream().AsChan()
	c2, _ := sc.Stream().AsChan()
	sc.Add("test")
	select {
	case <-time.After(1 * time.Millisecond):
		t.Fatal("no response")
	case data := <-c1:
		if data != "test" {
			t.Error("got wrong data")
		}
	}
	select {
	case <-time.After(1 * time.Millisecond):
		t.Fatal("no response")
	case data := <-c2:
		if data != "test" {
			t.Error("got wrong data")
		}
	}

}

func TestStreamFirst(t *testing.T) {
	sc := NewStreamController[string]()
	c := sc.Stream().First().AsChan()

	sc.Add("test")
	select {
	case <-time.After(1 * time.Second):
		t.Error("no response")
	case data := <-c:
		if data != "test" {
			t.Error("got wrong data")
		}
	}
}

func TestStreamFilter(t *testing.T) {
	sc := NewStreamController[int]()
	c, _ := sc.Stream().Where(func(d int) bool { return d != 2 }).AsChan()
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
			if data == 2 {
				t.Error("got 2")
			}
		}
	}
}

func TestStreamSplit(t *testing.T) {
	sc := NewStreamController[int]()
	y, n := sc.Stream().Split(func(d int) bool { return d != 2 })
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
			if data == 2 {
				t.Error("got 2")
			}
		}
	}
	for i := 0; i < 3; i++ {
		select {
		case <-time.After(1 * time.Millisecond):
			t.Fatal("no response")
		case data := <-c2:
			if data != 2 {
				t.Error("got 2")
			}
		}
	}
}

func TestStreamTransformer(t *testing.T) {
	sc := NewStreamController[int]()
	c, _ := TransformStream(sc.Stream(), func(d int) int { return d * 2 }).AsChan()
	sc.Add(5)
	select {
	case <-time.After(1 * time.Millisecond):
		t.Fatal("no response")
	case data := <-c:
		if data != 10 {
			t.Error("got 5")
		}
	}
}

func TestJoinFuture(t *testing.T) {

	sc := NewStreamController[string]()
	c, _ := sc.Stream().AsChan()
	fc := NewCompleter[string]()
	sc.JoinFuture(fc.Future())
	fc.Complete("test")
	select {
	case <-time.After(1 * time.Second):
		t.Error("no response")
	case data := <-c:
		if data != "test" {
			t.Error("got wrong data")
		}
	}
}
