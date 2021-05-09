package passthroughreader

import "io"

func NewPassthrough(r io.Reader, fn func(n int)) io.Reader {
	return passthroughImpl{r: r, fn: fn}
}

type passthroughImpl struct {
	r  io.Reader
	fn func(n int)
}

func (impl passthroughImpl) Read(b []byte) (int, error) {
	n, err := impl.r.Read(b)
	go impl.fn(n)
	return n, err
}
