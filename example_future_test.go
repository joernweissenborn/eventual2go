package eventual2go_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/joernweissenborn/eventual2go"
)

// Demonstrates the basic usage of futures
func ExampleFuture() {

	// create the completers, one we will complete with error, the other normaly.
	completerNor := eventual2go.NewCompleter[string]()
	completerErr := eventual2go.NewCompleter[string]()

	// set up success handler
	var onsuccess eventual2go.CompletionHandler[string] = func(d string) {
		fmt.Println("SUCESS:", d)
	}

	// set up error handler
	var onerror eventual2go.ErrorHandler = func(e error)  {
		fmt.Println("ERROR:", e)
	}

	// our long running async func
	mylongrunning := func(do_err bool, c *eventual2go.Completer[string]) {
		time.Sleep(1 * time.Second)

		if do_err {
			c.CompleteError(errors.New("Hello Future Error"))
		} else {
			c.Complete("Hello Future")
		}
	}

	// get the futures
	fNor := completerNor.Future()
	fErr := completerErr.Future()

	// register the handlers

	// we chain the succes
	fNor.Then(onsuccess)
	fNor.Err(onerror)
	fErr.Then(onsuccess)
	fErr.Err(onerror)

	// execute the functions
	go mylongrunning(false, completerNor)
	go mylongrunning(true, completerErr)

	// wait for futures to complete
	fNor.WaitUntilComplete()
	fErr.WaitUntilComplete()

	// everything is async, so the future is maybe complete, but the handlers must not have been executed necessarily, so we wait 10 ms
	time.Sleep(10 * time.Millisecond)
}
