package eventual2go

// Event represents a generic classifer assicated with generic data.
type Event struct {
	Classifier interface{}
	Data       Data
}
