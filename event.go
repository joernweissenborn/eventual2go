package eventual2go

// Event represents a generic classifer assicated with generic data.
type Event[T any] struct {
	Classifier interface{}
	Data       T
}
