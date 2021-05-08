package peerlist

import "example.com/gotorrent/lib/core/domain"

type PeerRepo interface {
	GetPeers() []domain.Host
}
