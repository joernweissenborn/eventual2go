package eventual2go_test

import (
	"testing"
	"time"

	"github.com/joernweissenborn/eventual2go"
)

func TestFutureWaitGroup(t *testing.T) {
	c1 := eventual2go.NewCompleter()
	c2 := eventual2go.NewCompleter()

	wg := eventual2go.NewFutureWaitGroup()
	wg.Add(c1.Future())
	wg.Add(c2.Future())

	c1.Complete(nil)
	c2.CompleteError(nil)

	c := make(chan interface{})
	go func() {
		wg.Wait()
		c <- nil
	}()

	select {
	case <-c:
	case <-time.After(10 * time.Millisecond):
		t.Error("Waited to long")
	}
}
