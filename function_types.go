package eventual2go

// A CompletionFunc is the argument for Completer.OnChange.
type CompletionFunc[T any] func() (T, error)

// A CompletionHandler gets invoked when a Future is completed. Returned value gets propagated when chaining futures.
type CompletionHandler[T any] func(T)

// An ErrorHandler gets invoked when a Future fails. Returned value and error are propagated when chaining futures, if
// the error is nil, the chained future will be completed with the data, otherwise it fails.
type ErrorHandler func(error)

// A Subscriber gets invoked whenever data is added to the consumed stream.
type Subscriber[T any] func(T)

// A DeriveSubscriber gets invoked every time data is added on the source stream and is responsible for adding the (transformed) data on the sink stream controller.
type DeriveSubscriber [T, V any] func(*StreamController[V], T)

// A Transformer gets invoked when data is added to the consumed stream. The output gets added to the transformed
// stream.
type Transformer[T, V any] func(T) V

// A TransformerConditional is like a Transformer, but can filter the data.
type TransformerConditional[T, V any] func(T) (V, bool)

// A Filter gets invoked when data is added to the consumed stream. The data is added to filtered stream conditionally,
// depending the Filter got registered with Where or WhereNot.
type Filter[T any] func(T) bool
