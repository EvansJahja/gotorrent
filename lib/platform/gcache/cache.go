package gcache

import (
	"errors"
	"sync"

	"example.com/gotorrent/lib/core/adapter/cache"

	"github.com/bluele/gcache"
)

func NewCache() cache.Cache {
	c := cacheImpl{
		fallbacks: &sync.Map{},
	}
	gc := gcache.New(10).LRU().LoaderFunc(c.loaderFunc).Build()
	c.gc = gc
	return c
}

type cacheImpl struct {
	gc gcache.Cache
	//fallbacks map[interface{}]func() (interface{}, error)
	fallbacks *sync.Map
}

func (c cacheImpl) loaderFunc(key interface{}) (interface{}, error) {
	v, ok := c.fallbacks.Load(key)
	if ok {
		w := v.(func() (interface{}, error))
		return w()
	}
	return nil, errors.New("no loader func")
}
func (c cacheImpl) Cached(key interface{}, fallback func() (interface{}, error)) (interface{}, error) {
	c.fallbacks.Store(key, fallback)

	return c.gc.Get(key)
	/*
		if err != nil {
			if errors.Is(err, gcache.KeyNotFoundError) {
				f, err := fallback()
				if err != nil {
					return nil, fmt.Errorf("%w error during fallback", err)
				}
				_ = c.gc.Set(key, f)
				return f, nil
			} else {
				return nil, errors.New("cache error")
			}
		}

	*/

}
