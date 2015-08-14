package eventual2go

import "sync"

type FutureCache struct {
	m         sync.Mutex
	indices   []int
	futures   []*Future
	nextindex int
	size      int
}

func NewCache(size int) (fc *FutureCache) {
	fc = new(FutureCache)
	fc.size = size
	fc.indices = make([]int, size)
	for i := range fc.indices {
		fc.indices[i] = -1
	}
	fc.futures = make([]*Future, size)
	return
}

func (fc FutureCache) Cached(Index int) (is bool) {
	for _, i := range fc.indices {
		is = i == Index
		if is {
			return
		}
	}
	return
}

func (fc FutureCache) Get(Index int) (f *Future) {
	for i, index := range fc.indices {
		if index == Index {
			f = fc.futures[i]
			return
		}
	}
	return
}

func (fc *FutureCache) incIndex() {
	fc.nextindex++
	if fc.nextindex == fc.size {
		fc.nextindex = 0
	}
}

func (fc *FutureCache) Cache(Index int, f *Future) {
	fc.indices[fc.nextindex] = Index
	fc.futures[fc.nextindex] = f
	fc.incIndex()
}
