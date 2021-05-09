// BROKEN
package fanreader

import (
	"io"
)

func NewFanReader(readers ...io.ReadSeeker) io.Reader {
	return &fanreadImpl{
		readers: readers,
	}
}

type fanreadImpl struct {
	readers []io.ReadSeeker
	cursor  int64
}

var _ io.Reader = &fanreadImpl{}

func (impl *fanreadImpl) Read(p []byte) (int, error) {
Retry:

	bCh := make(chan []byte, len(impl.readers))
	errCh := make(chan error, len(impl.readers))

	for _, r := range impl.readers {
		go func(r io.ReadSeeker) {
			gobyte := make([]byte, len(p))

			r.Seek(int64(impl.cursor), io.SeekStart)
			n, err := r.Read(gobyte)
			if err != nil {
				errCh <- err
				return
			}
			// success
			gobyte = gobyte[:n]
			bCh <- gobyte

		}(r)
	}

	var buf []byte
	var err error

	select {
	case b := <-bCh:
		buf = b
	case e := <-errCh:
		select {
		case b := <-bCh:
			buf = b
		default:
			err = e
		}
	}

	if err != nil {
		if err == io.EOF {
			return 0, err
		} else {
			goto Retry
		}
	}
	// success
	n := copy(p, buf)

	/*
		randReader.Seek(int64(impl.cursor), io.SeekStart)
		n, err := randReader.Read(p)
		if err != nil {
			if err != io.EOF {
				goto Retry
			}
		}
	*/
	impl.cursor += int64(n)

	return n, err
}
