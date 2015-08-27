package eventual2go

import (
	"testing"
	"time"
)

func TestStreamBasics(t *testing.T) {
	sc := NewStreamController()
	defer sc.Close()
	c := sc.AsChan()
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

func TestStreamClose(t *testing.T) {
	sc := NewStreamController()
	sc.Close()
	if !sc.Closed().Completed() {
		t.Error("stream didnt close")
	}
}

func TestStreamCancelSub(t *testing.T) {
	sc := NewStreamController()
	a := false
	ss := sc.Listen(func(d Data){
		a=true})
	ss.Close()
	sc.Add(0)
	time.Sleep(1*time.Millisecond)
	if a {
		t.Error("subscription didn't cancel")
	}
	if !ss.Closed().Completed() {
		t.Error("subscription future didn't complete")
	}

	b := false
	c := NewCompleter()
	sc.Listen(func(Data){b=true}).CloseOnFuture(c.Future())
	c.Complete(0)
	time.Sleep(1*time.Millisecond)

	sc.Add(0)
	time.Sleep(1*time.Millisecond)

	if b {
		t.Error("subscription didn't cancel")
	}

	ss = sc.Listen(func(d Data){
		a=true})
	sc.Close()
	time.Sleep(1*time.Millisecond)
	if !ss.Closed().Completed() {
		t.Error("subscription future didn't complete")
	}

}

func TestStreamMultiSubscription(t *testing.T) {
	sc := NewStreamController()
	defer sc.Close()
	c1 := sc.AsChan()
	c2 := sc.AsChan()
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
	defer sc.Close()
	c := sc.AsChan()

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
	defer sc.Close()
	c := sc.Where(func(d Data) bool { return d.(int) != 2 }).AsChan()
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
	defer sc.Close()
	y, n := sc.Split(func(d Data) bool { return d.(int) != 2 })
	c1 := y.AsChan()
	c2 := n.AsChan()
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
	defer sc.Close()
	c :=  sc.Transform(func(d Data) Data { return d.(int) * 2 }).AsChan()
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


