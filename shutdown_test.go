package eventual2go_test

import (
	"errors"
	"testing"

	"github.com/joernweissenborn/eventual2go"
)

func TestShutdown(t *testing.T) {

	sd := &testshutdown{}
	se := &testshutdownerr{}

	s := eventual2go.NewShutdown()

	s.Register(sd)
	s.Register(se)

	errs := s.Do(true)

	if !(sd.data.(bool)) {
		t.Error("wrong data")
	}

	if len(errs) != 1 {
		t.Error("wrong number of errors", len(errs))
	} else {
		if errs[0] != testerr {
			t.Error("wrong error")
		}
	}
}

type testshutdown struct {
	data eventual2go.Data
}

func (t *testshutdown) Shutdown(d eventual2go.Data) error {
	t.data = d
	return nil
}

var testerr = errors.New("test")

type testshutdownerr struct {
}

func (t *testshutdownerr) Shutdown(d eventual2go.Data) error {
	return testerr
}
