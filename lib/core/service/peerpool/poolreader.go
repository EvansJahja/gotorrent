package peerpool

import (
	"errors"
	"io"
	"math/rand"
	"time"

	"example.com/gotorrent/lib/core/adapter/peer"
	"example.com/gotorrent/lib/logger"
)

var l_poolreader = logger.Named("poolreader")

type poolReaderImpl struct {
	impl          *peerPoolImpl
	pieceNo       uint32
	pieceLength   int
	curSeek       int
	pieceCount    int
	torrentLength int
}

func (poolImpl *poolReaderImpl) Read(p []byte) (int, error) {
Retry:
	filteredPeers := FilterPool(poolImpl.impl.connectedPeers, FilterConnected, FilterNotChoking, FilterHasPiece(poolImpl.pieceNo))
	if len(filteredPeers) == 0 {
		time.Sleep(1 * time.Second)
		goto Retry
	}
	//peerIdx := 0
	peerIdx := rand.Int() % len(filteredPeers)
	targetPeer := filteredPeers[peerIdx]

	//fmt.Printf("Choosing %s out of %d peers\n", targetPeer.GetPeerID(), len(filteredPeers))

	//fmt.Println("Creating peer reader")

	// Limit reading to piece length
	pieceLength := poolImpl.getPieceLength()
	remainingToRead := pieceLength - poolImpl.curSeek
	bufClamp := len(p)
	if bufClamp > remainingToRead {
		bufClamp = remainingToRead
		if bufClamp < 0 {
			return 0, errors.New("invalid read pos")
		}
	}
	p = p[:bufClamp] // Clamp to remainingToRead

	// Todo: use sync.pool
	r := peer.NewPeerReader(targetPeer, poolImpl.pieceNo)
	r.Seek(int64(poolImpl.curSeek), io.SeekStart)
	//fmt.Println("Going to read")
	n, err := r.Read(p)
	//fmt.Println("Done reading")
	if err != nil {
		// l_poolreader.Sugar().Errorw("error reading", "err", err.Error(), "peerid", targetPeer.GetPeerID())
		return 0, err

	}
	poolImpl.curSeek += n
	if err != nil {
		panic("err is not nil")
	}
	if n == 0 {
		panic("n is 0")
	}
	return n, nil
}

func (poolImpl *poolReaderImpl) Seek(offset int64, whence int) (int64, error) {
	pieceLength := poolImpl.getPieceLength()

	switch whence {
	case io.SeekCurrent:
		poolImpl.curSeek += int(offset)
	case io.SeekStart:
		poolImpl.curSeek = int(offset)
	case io.SeekEnd:
		return 0, errors.New("seekend not implemented")
	}
	if poolImpl.curSeek > pieceLength {
		poolImpl.curSeek = pieceLength
	}
	return int64(poolImpl.curSeek), nil
}
func (poolImpl *poolReaderImpl) Close() error {
	// TODO
	return nil
}

func (poolImpl *poolReaderImpl) getPieceLength() int {

	pieceLength := poolImpl.pieceLength
	if poolImpl.pieceNo == uint32(poolImpl.pieceCount-1) {
		// Last
		lastRemainder := poolImpl.torrentLength % poolImpl.pieceLength
		if lastRemainder != 0 {
			pieceLength = lastRemainder
		}
	}
	return pieceLength
}
