package peer

import (
	"errors"
	"fmt"
	"io"
	"time"
)

type peerReader struct {
	pieceNo  uint32
	peer     Peer
	curPos   uint32
	dataChan chan []byte
}

func NewPeerReader(p Peer, pieceNo uint32) io.ReadSeekCloser {
	pr := peerReader{
		peer:     p,
		pieceNo:  pieceNo,
		dataChan: make(chan []byte),
	}
	return &pr

}

var _ io.ReadSeeker = &peerReader{}

func (r *peerReader) Read(p []byte) (n int, err error) {
	for r.peer.GetState().WeAreChocked {
		return 0, errors.New("choked")
	}

	requestedLength := uint32(len(p))

	if requestedLength <= 0 {
		return 0, io.EOF
	}

	fmt.Printf("Request #%x %x\n", r.pieceNo, r.curPos)
	//r.peer.RequestPiece(r.pieceNo, r.curPos, requestedLength)
	dataChan := r.peer.RequestPieceWithChan(r.pieceNo, r.curPos, requestedLength)
	select {
	case <-time.After(1 * time.Second):
		fmt.Printf("Timeout <%s> #%x %x\n", r.peer.Hostname(), r.pieceNo, r.curPos)
		return 0, errors.New("timeout waiting for piece")
	case recvData := <-dataChan:
		n = copy(p, recvData[:])
		fmt.Printf("Recv #%x %x\n", r.pieceNo, r.curPos)
		r.curPos += uint32(n)
		return
	}
}

func (r *peerReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		r.curPos = uint32(offset)
	case io.SeekCurrent:
		r.curPos += uint32(offset)
	case io.SeekEnd:
		return 0, errors.New("not supported")
	}
	return int64(r.curPos), nil
}

func (r *peerReader) Close() error {
	return nil
}
