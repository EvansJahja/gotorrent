package cache

type Cache interface {
	Cached(key interface{}, fallback func() (interface{}, error)) (interface{}, error)
}
