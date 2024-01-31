package eventual2go

type streamEvent [T any] struct {
	data T
	next *Future[*streamEvent[T]]
}
