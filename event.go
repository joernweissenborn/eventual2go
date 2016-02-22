package eventual2go

//go:generate event_generator -t Event

// Event represents generic data associated with an event name.
type Event struct {
	Name string
	Data Data
}
