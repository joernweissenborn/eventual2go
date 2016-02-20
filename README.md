[![Build Status](https://travis-ci.org/joernweissenborn/eventual2go.svg)](https://travis-ci.org/joernweissenborn/eventual2go)

# eventual2go

## Overview
A package for event-driven programming in Go.

Features:

* Streams
* Futures
* Reactor
* code generation for typed events

## Installation

Simply run

```
go get github.com/joernweissenborn/eventual2go
```

## Getting Started

eventual2go deals with cases, where you are in event-driven enviroment and want to react to something from which you don't now when its gonna happen.

Events can either be unique or not. For example, the result of reading in a large file is unique event, whereas a GET request on a webserver is not. eventual2go provides Futures for the former, Streams for the latter case.

An example of using futures for async filtereading looks like this.

```
```
