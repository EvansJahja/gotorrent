package peer

import (
	"example.com/gotorrent/lib/core/domain"
)

type State struct {
	WeAreChocked      bool
	TheyAreChocked    bool
	WeAreInterested   bool
	TheyAreInterested bool
	Connected         bool
}

type Peer interface {
	Connect() error
	GetHavePieces() map[int]struct{}

	GetMetadata() (domain.Metadata, error)
	Hostname() string

	RequestPiece(pieceId uint32, begin uint32, length uint32)
	RequestPieceWithChan(pieceId uint32, begin uint32, length uint32) <-chan []byte
	GetPeerID() []byte

	Choke()
	Unchoke()

	Interested()
	Uninterested()

	GetState() State

	OnChokedChanged(func(isChoked bool))
	OnPiecesUpdatedChanged(func())
	OnPieceArrive(func(index uint32, begin uint32, piece []byte))

	DisconnectOnChokedChanged(func(isChoked bool))
	DisconnectOnPiecesUpdatedChanged(func())
	DisconnectOnPieceArrive(func(index uint32, begin uint32, piece []byte))

	PieceRequests() <-chan PieceRequest
}

type PieceRequest struct {
	PieceNo  uint32
	Begin    uint32
	Length   uint32
	Response chan<- []byte
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
