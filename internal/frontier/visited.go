package frontier

import (
	"hash/fnv"
	"sync"
)

type Visited struct{
	set map[uint64]struct{}
	mu sync.Mutex
}

func NewVisited() *Visited{
	return &Visited{
		set: make(map[uint64] struct{}),
	}
}

func hash(s string) uint64{
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func (v*Visited) Add(u string){
	v.mu.Lock()
	defer v.mu.Unlock()
	v.set[hash(u)] = struct{}{}
}

func (v *Visited) Has(u string) bool{
	v.mu.Lock()
	defer v.mu.Unlock()
	_, ok := v.set[hash(u)]
	return ok
}

func (v *Visited) Size() int{
	v.mu.Lock()
	defer v.mu.Unlock()
	return len(v.set)
}

