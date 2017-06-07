package eventual2go

import "testing"

func TestObservable(t *testing.T) {
	o := NewObservable("foo")
	if o.Value().(string) != "foo" {
		t.Fatal("Wrong initial Value")
	}

	testvals := []int{5, 9, 2, 5, 4}
	c, _ := o.AsChan()
	for _, v := range testvals {
		o.Change(v)
	}

	for _, v := range testvals {
		cv := <-c
		if cv.(int) != v {
			t.Fatalf("Wrong Change Value, Want %d, have %d", v, o.Value())
		}

	}
}
