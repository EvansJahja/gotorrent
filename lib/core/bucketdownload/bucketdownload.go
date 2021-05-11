package bucketdownload

import (
	"fmt"
	"io"
	"math"
	"sync"
)

func New(src func() io.ReadSeekCloser, sink func() io.WriteSeeker, chunkSize int, totalSize int, concurrentWorker int) Bucket {
	return &impl{
		src:              src,
		sink:             sink,
		chunkSize:        chunkSize,
		totalSize:        totalSize,
		concurrentWorker: concurrentWorker,
	}

}

type Bucket interface {
	Start()
}

type impl struct {
	src              func() io.ReadSeekCloser
	sink             func() io.WriteSeeker
	chunkSize        int
	totalSize        int
	concurrentWorker int
}

func (impl *impl) Start() {

	goroutinelim := make(chan struct{}, impl.concurrentWorker)
	var wg sync.WaitGroup

	for i := 0; i < impl.numOfChunks(); i++ {
		wg.Add(1)
		go func(chunkNo int) {
			goroutinelim <- struct{}{}
			start, size := impl.rangeFromChunkNo(chunkNo)
			reader := impl.src()
			writer := impl.sink()

			reader.Seek(int64(start), io.SeekStart)
			writer.Seek(int64(start), io.SeekStart)

			limReader := io.LimitReader(reader, int64(size))
		Retry:
			n, err := io.Copy(writer, limReader)
			if err != nil {
				if n != 0 {
					panic("n is not 0")
				}
				if err != io.EOF {
					goto Retry
				}
			}
			if n == 0 {
				fmt.Println("n is 0")
			}
			reader.Close()
			wg.Done()
			<-goroutinelim
		}(i)
	}
	wg.Wait()

}
func (impl *impl) numOfChunks() int {
	return int(math.Ceil(float64(impl.totalSize) / float64(impl.chunkSize)))
}
func (impl *impl) rangeFromChunkNo(chunkNo int) (start, size int) {
	return chunkNo * impl.chunkSize, impl.chunkSize

}

// from peerpools, we can download limited amount of bytes
// we need to save the result to some sort of writer, using buckets
// need to retry

// interface:
// after connecting source and sink, sink needs to be divided into buckets.
// manage whether bucket is successfully filled or not (and needing retry)
