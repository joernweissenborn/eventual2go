package eventual2go

// A CompletionHandler gets invoked when a Future is completed. Returned value gets propagated when chaining futures.
type CompletionHandler func(Data) Data

// An ErrorHandler gets invoked when a Future fails. Returned value and error are propagated when chaining futures, if
// the error is nil, the chained future will be completed with the data, otherwise it fails.
type ErrorHandler func(error) (Data, error)

// A Subscriber gets invoked whenever data is added to the consumed stream.
type Subscriber func(Data)

// A Subscriber gets invoked whenever data is added to the consumed stream.
type DeriveSubscriber func(*StreamController, Data)

// A Transformer gets invoked when data is added to the consumed stream. The output gets added to the transformed
// stream.
type Transformer func(Data) Data

// A Filter gets invoked when data is added to the consumed stream. The data is added to filtered stream conditionally,
// depending the Filter got registered with Where or WhereNot.
type Filter func(Data) bool
