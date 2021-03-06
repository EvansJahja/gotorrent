//go:generate mockgen -destination ../../../mocks/peer/peer.go . Peer

package peer

import (
	"example.com/gotorrent/lib/core/domain"
	"example.com/gotorrent/lib/extensions"
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
	TheirPieces() domain.PieceList
	OurPieces() domain.PieceList
	SetOurPiece(pieceNo uint32)

	GetMetadata() (domain.Metadata, error)
	Hostname() string

	RequestPiece(pieceId uint32, begin uint32, length uint32) <-chan []byte

	// Not actually peer ID, but hashed version (easier to log and carry)
	GetID() string

	Choke()
	Unchoke()

	Interested()
	Uninterested()

	GetState() State

	GetDownloadRate() float32
	GetUploadRate() float32
	GetDownloadBytes() uint32
	GetUploadBytes() uint32
	TellPieceCompleted(pieceNo uint32)

	OnChokedChanged(func(isChoked bool))
	OnPiecesUpdatedChanged(func())

	DisconnectOnChokedChanged(func(isChoked bool))
	DisconnectOnPiecesUpdatedChanged(func())

	PieceRequests() <-chan PieceRequest
	ExtHandler() extensions.ExtHandler
}

type PieceRequest struct {
	PieceNo  uint32
	Begin    uint32
	Length   uint32
	Response chan<- []byte
}
