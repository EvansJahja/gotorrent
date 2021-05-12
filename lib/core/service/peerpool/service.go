package peerpool

import (
	"io"

	"example.com/gotorrent/lib/core/domain"
)

type Service interface {
	AddHosts(newHosts ...domain.Host)
	Start()
	//	NewPeerPoolReader(pieceNo uint32, pieceLength uint32) io.ReadSeekCloser
	NewPeerPoolReader(pieceNo uint32, pieceLength int, pieceCount int, torrentLength int) io.ReadSeekCloser
}

var _ Service = &Impl{}
