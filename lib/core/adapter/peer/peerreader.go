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
	isChoked bool
	curPos   uint32
	dataChan chan []byte
}

func NewPeerReader(p Peer, pieceNo uint32) io.ReadSeekCloser {
	pr := peerReader{
		peer:     p,
		pieceNo:  pieceNo,
		dataChan: make(chan []byte),
		isChoked: true,
	}
	p.OnChokedChanged(pr.onChokedChanged)
	p.OnPieceArrive(pr.onPieceArrive)
	return &pr

}

var _ io.ReadSeeker = &peerReader{}

func (r *peerReader) onChokedChanged(isChoked bool) {
	r.isChoked = isChoked
}
func (r *peerReader) onPieceArrive(index uint32, begin uint32, data []byte) {
	if r.pieceNo != index {
		// not for me
		return
	}
	if begin != uint32(r.curPos) {
		// Ignore
		return
	}

	//data = data[r.curPos:] /// trim data we already seen

	r.dataChan <- data

}
func (r *peerReader) Read(p []byte) (n int, err error) {
	for r.peer.GetState().WeAreChocked {
		return 0, errors.New("choked")
	}

	requestedLength := uint32(len(p))

	if requestedLength <= 0 {
		return 0, io.EOF
	}

	fmt.Printf("Request %d %d\n", r.curPos, requestedLength)
	r.peer.RequestPiece(r.pieceNo, r.curPos, requestedLength)
	select {
	case <-time.After(5 * time.Second):
		fmt.Printf("Timeout %d %d\n", r.curPos, requestedLength)
		return 0, errors.New("timeout waiting for piece")
	case recvData := <-r.dataChan:
		n = copy(p, recvData[:])
		fmt.Printf("Recv %d %d\n", r.curPos, n)
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
	r.peer.DisconnectOnChokedChanged(r.onChokedChanged)
	r.peer.DisconnectOnPieceArrive(r.onPieceArrive)
	return nil
}
