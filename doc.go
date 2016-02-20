/*
Package eventual2go is a library for event driven development. It provides implementations for futures and streams, as can be found in many modern languages.

eventual2go deals with cases, where you are in event-driven enviroment and want to react to something from which you don't now when its gonna happen.

Events can either be unique or not. For example, the result of reading in a large file is unique event, whereas a GET request on a webserver is not. eventual2go provides Futures for the former, Streams for the latter case.
*/
package eventual2go
