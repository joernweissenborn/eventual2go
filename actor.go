package eventual2go

// Message is a classifier for a message received event.
type Message struct{}

// ActorMessageStream is used to send messages to an actor.
type ActorMessageStream struct {
	streamController *StreamController
	shutdown         *Completer
}

func newActorMessageStream() (ams ActorMessageStream) {
	ams = ActorMessageStream{
		streamController: NewStreamController(),
		shutdown:         NewCompleter(),
	}
	return
}

// Send sends a message to an actor.
func (ams ActorMessageStream) Send(data Data) {
	ams.streamController.Add(data)
}

// Shutdown sends a shutdown signal to the actor. Messages send before the shutdown signal are guaranteed to be handled.
func (ams ActorMessageStream) Shutdown(data Data) {
	ams.shutdown.Complete(data)
}

// Actor is a simple actor.
type Actor interface {
	Init() error
	OnMessage(d Data)
	Shutdown(d Data)
}

// SpawnActor creates an actor and returns a message stream to it.
func SpawnActor(a Actor) (messages ActorMessageStream, err error) {

	messages = newActorMessageStream()

	if err = a.Init(); err != nil {
		return
	}

	actor := NewReactor()
	actor.React(Message{}, a.OnMessage)
	actor.AddStream(Message{}, messages.streamController.Stream())
	actor.AddFuture(ShutdownEvent{}, messages.shutdown.Future())
	actor.OnShutdown(a.Shutdown)

	return
}
