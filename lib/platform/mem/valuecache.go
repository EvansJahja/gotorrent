package mem

import (
	"sync"

	"example.com/gotorrent/lib/core/adapter/temprepo"
)

type MemValueCache struct {
	SyncMap *sync.Map
}

var _ temprepo.TempMetadata = MemValueCache{}

func (m MemValueCache) Get(key interface{}) (val interface{}, ok bool) {
	return m.SyncMap.Load(key)
}

func (m MemValueCache) Set(key interface{}, val interface{}) {
	m.SyncMap.Store(key, val)
}
