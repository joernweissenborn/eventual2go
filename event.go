package eventual2go

//go:generate event_generator -t Event

// Event represents a generic classifer assicated with generic data.
type Event struct {
	Classifier interface{}
	Data Data
}
