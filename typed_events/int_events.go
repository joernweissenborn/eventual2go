
/*
 * generated by event_generator
 *
 * DO NOT EDIT
 */

package typed_events

import "github.com/joernweissenborn/eventual2go"



type IntCompleter struct {
	*eventual2go.Completer
}

func NewIntCompleter() *IntCompleter {
	return &IntCompleter{eventual2go.NewCompleter()}
}

func (c *IntCompleter) Complete(d int) {
	c.Completer.Complete(d)
}

func (c *IntCompleter) Future() *IntFuture {
	return &IntFuture{c.Completer.Future()}
}

type IntFuture struct {
	*eventual2go.Future
}

func (f *IntFuture) GetResult() int {
	return f.Future.GetResult().(int)
}

type IntCompletionHandler func(int) int

func (ch IntCompletionHandler) toCompletionHandler() eventual2go.CompletionHandler {
	return func(d eventual2go.Data) eventual2go.Data {
		return ch(d.(int))
	}
}

func (f *IntFuture) Then(ch IntCompletionHandler) *IntFuture {
	return &IntFuture{f.Future.Then(ch.toCompletionHandler())}
}

func (f *IntFuture) AsChan() chan int {
	c := make(chan int, 1)
	cmpl := func(d chan int) IntCompletionHandler {
		return func(e int) int {
			d <- e
			close(d)
			return e
		}
	}
	ecmpl := func(d chan int) eventual2go.ErrorHandler {
		return func(error) (eventual2go.Data, error) {
			close(d)
			return nil, nil
		}
	}
	f.Then(cmpl(c))
	f.Err(ecmpl(c))
	return c
}

type IntStreamController struct {
	*eventual2go.StreamController
}

func NewIntStreamController() *IntStreamController {
	return &IntStreamController{eventual2go.NewStreamController()}
}

func (sc *IntStreamController) Add(d int) {
	sc.StreamController.Add(d)
}

func (sc *IntStreamController) Join(s *IntStream) {
	sc.StreamController.Join(s.Stream)
}

func (sc *IntStreamController) JoinFuture(f *IntFuture) {
	sc.StreamController.JoinFuture(f.Future)
}

func (sc *IntStreamController) Stream() *IntStream {
	return &IntStream{sc.StreamController.Stream()}
}

type IntStream struct {
	*eventual2go.Stream
}

type IntSubscriber func(int)

func (l IntSubscriber) toSubscriber() eventual2go.Subscriber {
	return func(d eventual2go.Data) { l(d.(int)) }
}

func (s *IntStream) Listen(ss IntSubscriber) *eventual2go.Subscription {
	return s.Stream.Listen(ss.toSubscriber())
}

type IntFilter func(int) bool

func (f IntFilter) toFilter() eventual2go.Filter {
	return func(d eventual2go.Data) bool { return f(d.(int)) }
}

func toIntFilterArray(f ...IntFilter) (filter []eventual2go.Filter){

	filter = make([]eventual2go.Filter, len(f))
	for i, el := range f {
		filter[i] = el.toFilter()
	}
	return
}

func (s *IntStream) Where(f ...IntFilter) *IntStream {
	return &IntStream{s.Stream.Where(toIntFilterArray(f...)...)}
}

func (s *IntStream) WhereNot(f ...IntFilter) *IntStream {
	return &IntStream{s.Stream.WhereNot(toIntFilterArray(f...)...)}
}

func (s *IntStream) Split(f IntFilter) (*IntStream, *IntStream)  {
	return s.Where(f), s.WhereNot(f)
}

func (s *IntStream) First() *IntFuture {
	return &IntFuture{s.Stream.First()}
}

func (s *IntStream) FirstWhere(f... IntFilter) *IntFuture {
	return &IntFuture{s.Stream.FirstWhere(toIntFilterArray(f...)...)}
}

func (s *IntStream) FirstWhereNot(f ...IntFilter) *IntFuture {
	return &IntFuture{s.Stream.FirstWhereNot(toIntFilterArray(f...)...)}
}

func (s *IntStream) AsChan() (c chan int) {
	c = make(chan int)
	s.Listen(pipeToIntChan(c)).Closed().Then(closeIntChan(c))
	return
}

func pipeToIntChan(c chan int) IntSubscriber {
	return func(d int) {
		c <- d
	}
}

func closeIntChan(c chan int) eventual2go.CompletionHandler {
	return func(d eventual2go.Data) eventual2go.Data {
		close(c)
		return nil
	}
}

type IntCollector struct {
	*eventual2go.Collector
}

func NewIntCollector() *IntCollector {
	return &IntCollector{eventual2go.NewCollector()}
}

func (c *IntCollector) Add(d int) {
	c.Collector.Add(d)
}

func (c *IntCollector) AddFuture(f *IntFuture) {
	c.Collector.Add(f.Future)
}

func (c *IntCollector) AddStream(s *IntStream) {
	c.Collector.AddStream(s.Stream)
}

func (c *IntCollector) Get() int {
	return c.Collector.Get().(int)
}

func (c *IntCollector) Preview() int {
	return c.Collector.Preview().(int)
}
