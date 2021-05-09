package fanreader

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
)

type NopSeeker struct{ io.Reader }

func (NopSeeker) Seek(_ int64, _ int) (int64, error) {
	return 0, nil
}

func Test(t *testing.T) {
	b := make([]byte, 1000)

	r1 := bytes.NewReader(b)
	r2 := bytes.NewReader(b)
	r3 := NopSeeker{iotest.ErrReader(errors.New("aaa"))}

	rfan := NewFanReader(r1, r2, r3)
	err := iotest.TestReader(rfan, b)
	assert.NoError(t, err)

}
