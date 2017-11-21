package eventual2go

// A CompletionFunc is the argument for Completer.OnChange.
type CompletionFunc func() (Data, error)

// A CompletionHandler gets invoked when a Future is completed. Returned value gets propagated when chaining futures.
type CompletionHandler func(Data) Data

// An ErrorHandler gets invoked when a Future fails. Returned value and error are propagated when chaining futures, if
// the error is nil, the chained future will be completed with the data, otherwise it fails.
type ErrorHandler func(error) (Data, error)

// A Subscriber gets invoked whenever data is added to the consumed stream.
type Subscriber func(Data)

// A DeriveSubscriber gets invoked every time data is added on the source stream and is responsible for adding the (transformed) data on the sink stream controller.
type DeriveSubscriber func(*StreamController, Data)

// A Transformer gets invoked when data is added to the consumed stream. The output gets added to the transformed
// stream.
type Transformer func(Data) Data

// A TransformerConditional is like a Transformer, but can filter the data.
type TransformerConditional func(Data) (Data, bool)

// A Filter gets invoked when data is added to the consumed stream. The data is added to filtered stream conditionally,
// depending the Filter got registered with Where or WhereNot.
type Filter func(Data) bool
