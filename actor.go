package eventual2go

// Message is a classifier for a message received event.
type Message struct{}

// ActorMessageStream is used to send messages to an actor.
type ActorMessageStream struct {
	streamController *StreamController
	shutdown         *Completer
	finalErr         *Future
}

func newActorMessageStream(finalErr *Future) (ams ActorMessageStream) {
	ams = ActorMessageStream{
		streamController: NewStreamController(),
		shutdown:         NewCompleter(),
		finalErr:         finalErr,
	}
	return
}

// Send sends a message to an actor.
func (ams ActorMessageStream) Send(data Data) {
	ams.streamController.Add(data)
}

// Shutdown sends a shutdown signal to the actor. Messages send before the shutdown signal are guaranteed to be handled.
func (ams ActorMessageStream) Shutdown(data Data) (err error) {
	ams.shutdown.Complete(data)
	err = (<-ams.finalErr.AsChan()).(error)
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

func shutdownHandler(finalErr *Completer, a Actor) Subscriber {
	return func(d Data) {
		var err error
		if s, ok := a.(ShutdownActor); ok {
			err = s.Shutdown(d)
		}
		finalErr.Complete(err)
	}
}

// LoopActor is an actor with a loop method which is called repeatedly. Messages are handled in between loop repetitions.
type LoopActor interface {
	Actor
	Loop() (cont bool)
}

type loopEvent struct{}

func loopHandler(r *Reactor, la LoopActor) Subscriber {
	return func(d Data) {
		if la.Loop() {
			r.Fire(loopEvent{}, nil) // if the actor is should down alredy, Fire will do nothing
		}
	}
}

// SpawnActor creates an actor and returns a message stream to it.
func SpawnActor(a Actor) (messages ActorMessageStream, err error) {

	finalErr := NewCompleter()
	messages = newActorMessageStream(finalErr.Future())

	if err = a.Init(); err != nil {
		return
	}

	actor := NewReactor()
	actor.React(Message{}, a.OnMessage)
	actor.AddStream(Message{}, messages.streamController.Stream())
	actor.AddFuture(ShutdownEvent{}, messages.shutdown.Future())
	actor.OnShutdown(shutdownHandler(finalErr, a))
	if la, ok := a.(LoopActor); ok {
		actor.React(loopEvent{}, loopHandler(actor, la))
		actor.Fire(loopEvent{}, nil)
	}

	return
}
