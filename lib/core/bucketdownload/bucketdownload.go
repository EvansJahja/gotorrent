package bucketdownload

import (
	"fmt"
	"io"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/context"
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
	Error() error
	Wait()
	Stats() Status
}

type Status struct {
	PieceNo uint32

	// Buckets
	Waiting  int32
	Starting int32
	Finished int32

	Progress float32
}

type impl struct {
	src              func() io.ReadSeekCloser
	sink             func() io.WriteSeeker
	chunkSize        int
	totalSize        int
	concurrentWorker int
	doneWg           sync.WaitGroup
	doneErr          error

	waiting     int32
	starting    int32
	finished    int32
	bucketcount int32
}

func (impl *impl) Stats() Status {
	s := Status{
		Waiting:  atomic.LoadInt32(&impl.waiting),
		Starting: atomic.LoadInt32(&impl.starting),
		Finished: atomic.LoadInt32(&impl.finished),
	}

	s.Progress = 100 * float32(s.Finished) / float32(impl.bucketcount)

	return s

}
func (impl *impl) Wait() {
	impl.doneWg.Wait()
}

func (impl *impl) Start() {
	impl.doneWg.Add(1) // to track that we are starting
	go impl.run()
}
func (impl *impl) Error() error {
	return impl.doneErr
}

func (impl *impl) run() {
	impl.doneErr = nil

	goroutinelim := make(chan struct{}, impl.concurrentWorker)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Minute)

	impl.bucketcount = int32(impl.numOfChunks())
	for i := 0; i < impl.numOfChunks(); i++ {
		if ctx.Err() != nil {
			break
		}
		atomic.AddInt32(&impl.waiting, 1)
		impl.doneWg.Add(1)
		go func(chunkNo int) {
			goroutinelim <- struct{}{}
			atomic.AddInt32(&impl.waiting, -1)
			atomic.AddInt32(&impl.starting, 1)
			start, size := impl.rangeFromChunkNo(chunkNo)
			reader := impl.src()
			writer := impl.sink()

			reader.Seek(int64(start), io.SeekStart)
			writer.Seek(int64(start), io.SeekStart)

			limReader := io.LimitReader(reader, int64(size))
		Retry:
			if ctx.Err() != nil {
				impl.doneErr = ctx.Err()
				impl.doneWg.Done()
				<-goroutinelim
				return
			}
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
			atomic.AddInt32(&impl.starting, -1)
			atomic.AddInt32(&impl.finished, 1)
			reader.Close()
			impl.doneWg.Done()
			<-goroutinelim
		}(i)
	}
	impl.doneErr = ctx.Err()

	impl.doneWg.Done() // decrement the one from start
	//return ctx.Err()
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
