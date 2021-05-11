package bucketdownload

import (
	"bytes"
	cryptrand "crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type BytesWriteSeeker struct {
	Buff   []byte
	cursor int64
}

func (b *BytesWriteSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		b.cursor += offset
	case io.SeekStart:
		b.cursor = offset
	case io.SeekEnd:
		b.cursor = int64(len(b.Buff)) + offset
	}
	return b.cursor, nil
}

func (b *BytesWriteSeeker) Write(p []byte) (n int, err error) {
	n = copy(b.Buff[b.cursor:], p)
	return
}

type NopReadSeekCloser struct {
	prev io.ReadSeeker
}

func (n NopReadSeekCloser) Read(p []byte) (int, error) {
	time.Sleep(time.Duration(len(p)) * time.Millisecond) // Intentionally make it slow
	if rand.Float32() < 0.4 {
		return 0, errors.New("error")
	}
	return n.prev.Read(p)
}

func (n NopReadSeekCloser) Seek(offset int64, whence int) (int64, error) {
	return n.prev.Seek(offset, whence)
}

func (n NopReadSeekCloser) Close() error {
	return nil
}

func Test(t *testing.T) {

	srcBuf := bytes.Buffer{}
	io.CopyN(&srcBuf, cryptrand.Reader, 1000)
	//src :=

	bb := make([]byte, 1000)

	src := func() io.ReadSeekCloser {
		r := bytes.NewReader(srcBuf.Bytes())
		return NopReadSeekCloser{prev: r}
	}

	dst := func() io.WriteSeeker {
		bws := BytesWriteSeeker{Buff: bb}
		return &bws
	}

	start := time.Now()
	b := New(src, dst, 100, 1000, 5)
	b.Start()
	dur := time.Since(start)

	assert.True(t, bytes.Equal(srcBuf.Bytes(), bb))
	assert.Less(t, dur, 2*time.Second, "should be fast")

	fmt.Println("Done")
}
