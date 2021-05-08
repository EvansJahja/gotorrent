package peer

import (
	"io"

	"example.com/gotorrent/lib/core/domain"
)

type Peer interface {
	Connect() error
	GetHavePieces() map[int]struct{}

	RequestPiece(pieceId int) io.Reader
}

type PeerFactory interface {
	New(h domain.Host) Peer
}

type PeerFactoryFn func(h domain.Host) Peer

func (p PeerFactoryFn) New(h domain.Host) Peer {
	return p(h)
}
