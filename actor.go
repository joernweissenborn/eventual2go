package eventual2go

type message struct {
	data Data
}

type loop struct{}
type shutdown message

//ActorMessageStream is used to send messages to an actor.
type ActorMessageStream struct {
	streamController *StreamController
	finalErr         *Future
}

func newActorMessageStream(finalErr *Future) (ams ActorMessageStream) {
	ams = ActorMessageStream{
		streamController: NewStreamController(),
		finalErr:         finalErr,
	}
	return
}

// Send sends a message to an actor.
func (ams ActorMessageStream) Send(data Data) {
	ams.streamController.Add(message{data})
}

// Shutdown sends a shutdown signal to the actor. Messages send before the shutdown signal are guaranteed to be handled.
func (ams ActorMessageStream) Shutdown(data Data) (err error) {
	ams.streamController.Add(shutdown{data})
	ferr := (<-ams.finalErr.AsChan())
	if ferr != nil {
		return ferr.(error)
	}
	return
}

// Actor is a simple actor.
type Actor interface {
	Init() error
	OnMessage(d Data)
}

// ShutdownActor is an actor with a Shutdown method, which is called upon actor shutdown.
type ShutdownActor interface {
	Actor
	Shutdown(Data) error
}

// LoopActor is an actor with a loop method which is called repeatedly. Messages are handled in between loop repetitions.
type LoopActor interface {
	Actor
	Loop() (cont bool)
}


// SpawnActor creates an actor and returns a message stream to it.
func SpawnActor(a Actor) (messages ActorMessageStream, err error) {

	if err = a.Init(); err != nil {
		return
	}

	finalErr := NewCompleter()
	messages = newActorMessageStream(finalErr.Future())
	messages.streamController.Stream().Listen(messageHandler(a, messages.streamController, finalErr))

	if _, ok := a.(LoopActor); ok {
		messages.streamController.Add(loop{})
	}

	return
}

func messageHandler(a Actor, msg *StreamController, finalErr *Completer) Subscriber {
	return func(d Data) {
		switch d.(type) {
		case message:
			a.OnMessage(d.(message).data)
		case loop:
			a.(LoopActor).Loop()
			msg.Add(loop{})
		case shutdown:
			var err error
			if s, ok := a.(ShutdownActor); ok {
				err = s.Shutdown(d.(shutdown).data)
			}
			finalErr.Complete(err)

		}
	}
}
