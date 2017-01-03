package eventual2go

type streamEvent struct {
	data Data
	next *Future
}
