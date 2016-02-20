package eventual2go

import "sync"

// FutureCache is a thread-safe cache for storing futures. It stores data with a userdefined index. Useful e.g. when needing to retrieve the same data for multiple requests from a slow location. The cache is sized and implemented as a ring buffer.
type FutureCache struct {
	m         sync.Mutex
	indices   []int
	futures   []*Future
	nextindex int
	size      int
}

// NewCache creates a new FutureCache of the given size
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

// Cached indicates if there has already been a Future cached at given index.
func (fc *FutureCache) Cached(index int) (is bool) {
	for _, i := range fc.indices {
		is = i == index
		if is {
			return
		}
	}
	return
}

// Get retrives the future with a given index.
func (fc *FutureCache) Get(index int) (f *Future) {
	for i, ind := range fc.indices {
		if index == ind {
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

// Cache stores a future with given index.
func (fc *FutureCache) Cache(Index int, f *Future) {
	fc.indices[fc.nextindex] = Index
	fc.futures[fc.nextindex] = f
	fc.incIndex()
}
