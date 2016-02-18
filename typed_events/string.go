package typed_events

import "github.com/joernweissenborn/eventual2go"

type StringCompleter struct {
	*eventual2go.Completer
}

func (c *StringCompleter) Complete(d string) {
	c.Completer.Complete(d)
}

func (c *StringCompleter) Future() *StringFuture {
	return &StringFuture{c.Completer.Future()}
}

type StringFuture struct {
	*eventual2go.Future
}

type StringCompletionHandler func(string) string

func (ch StringCompletionHandler) toCompletionHandler() eventual2go.CompletionHandler {
	return func(d eventual2go.Data) eventual2go.Data {
		return ch(d.(string))
	}
}

func (f *StringFuture) Then(ch StringCompletionHandler) *StringFuture {
	return &StringFuture{f.Future.Then(ch.toCompletionHandler())}
}

type StringStream struct {
	*eventual2go.Stream
}

type StringSuscriber func(string)

func (l StringSuscriber) toSuscriber() eventual2go.Subscriber {
	return func(d eventual2go.Data) { l(d.(string)) }
}

func (s *StringStream) Listen(ss StringSuscriber) *eventual2go.Subscription{
	return s.Stream.Listen(ss.toSuscriber())
}

type StringFilter func(string) bool

func (f StringFilter) toFilter() eventual2go.Filter {
	return func(d eventual2go.Data) bool { return f(d.(string)) }
}

func (s *StringStream) Where(f StringFilter) {
	s.Stream.Where(f.toFilter())
}

func (s *StringStream) WhereNot(f StringFilter) {
	s.Stream.WhereNot(f.toFilter())
}

func (s *StringStream) First() *StringFuture {
	return &StringFuture{s.Stream.First()}
}

func (s *StringStream) FirstWhere(f StringFilter) *StringFuture {
	return &StringFuture{s.Stream.FirstWhere(f.toFilter())}
}

func (s *StringStream) FirstWhereNot(f StringFilter) *StringFuture {
	return &StringFuture{s.Stream.FirstWhereNot(f.toFilter())}
}

func (s *StringStream) AsChan() (c chan string) {
	c = make(chan string)
	s.Listen(pipeToChan(c)).Closed().Then(closeChan(c))
	return
}

func pipeToChan(c chan string) StringSuscriber {
	return func(d string) {
		c<-d
	}
}

func closeChan(c chan string) eventual2go.CompletionHandler {
	return func(d eventual2go.Data) eventual2go.Data {
		close(c)
		return nil
	}
}

