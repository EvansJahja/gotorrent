package dummy

import (
	"example.com/gotorrent/lib/core/adapter/peerlist"
	"example.com/gotorrent/lib/core/domain"
)

func PeerList() peerlist.PeerRepo {
	return dummyPeerList{}
}

type dummyPeerList struct{}

var _ peerlist.PeerRepo = dummyPeerList{}

func (dummyPeerList) GetPeers() []domain.Host {
	return []domain.Host{
		{
			IP:   []byte{1, 2, 3, 0},
			Port: 1001,
		},
		{
			IP:   []byte{1, 2, 3, 1},
			Port: 1002,
		},
	}
}
