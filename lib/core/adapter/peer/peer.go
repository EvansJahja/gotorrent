package peer

import (
	"example.com/gotorrent/lib/core/domain"
)

type Peer interface {
	Connect() error
	GetHavePieces() map[int]struct{}

	GetMetadata() (domain.Metadata, error)

	RequestPiece(pieceId int, begin int, length int)
	GetPeerID() []byte

	Choke()
	Unchoke()
	Interested()
	Uninterested()

	OnChokedChanged(func(isChoked bool))
	OnPiecesUpdatedChanged(func())
	OnPieceArrive(func(index uint32, begin uint32, piece []byte))
}

type PeerFactory interface {
	New(h domain.Host) Peer
}

type PeerFactoryWithHashFn func(h domain.Host, infoHash []byte) Peer

func NewPeerFactory(infoHash []byte, peerFactoryWithHashFn PeerFactoryWithHashFn) PeerFactory {
	return peerFactoryFn(
		func(h domain.Host) Peer {
			return peerFactoryWithHashFn(h, infoHash)
		})
}

type peerFactoryFn func(h domain.Host) Peer

func (p peerFactoryFn) New(h domain.Host) Peer {
	return p(h)
}
