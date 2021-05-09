package peer

import (
	"errors"
	"io"
	"time"
)

type peerReader struct {
	pieceNo     uint32
	pieceLength uint32
	peer        Peer
	isChoked    bool
	curPos      uint32
	dataChan    chan []byte
}

func NewPeerReader(p Peer, pieceNo uint32, pieceLength uint32) io.ReadSeeker {
	pr := peerReader{
		peer:        p,
		pieceNo:     pieceNo,
		pieceLength: pieceLength,
		dataChan:    make(chan []byte),
		isChoked:    true,
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
	for r.isChoked {
		return 0, errors.New("choked")
	}

	requestedLength := uint32(len(p))

	maxLengthToRequest := r.pieceLength - r.curPos
	if requestedLength > maxLengthToRequest {
		requestedLength = maxLengthToRequest
	}

	if requestedLength <= 0 {
		return 0, io.EOF
	}

	r.peer.RequestPiece(r.pieceNo, r.curPos, requestedLength)
	select {
	case <-time.After(5 * time.Second):
		return 0, errors.New("timeout waiting for piece")
	case recvData := <-r.dataChan:
		n = copy(p, recvData[:])
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
