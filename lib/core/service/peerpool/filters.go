package peerpool

import "example.com/gotorrent/lib/core/adapter/peer"

func FilterNotChoking(peers []peer.Peer) []peer.Peer {
	return filter(peers, func(p peer.Peer) bool {
		return !p.GetState().WeAreChocked
	})
}

func FilterConnected(peers []peer.Peer) []peer.Peer {
	return filter(peers, func(p peer.Peer) bool {
		return p.GetState().Connected
	})

}

func FilterHasPiece(pieceNo uint32) func(peers []peer.Peer) []peer.Peer {
	return func(peers []peer.Peer) []peer.Peer {
		return filter(peers, func(p peer.Peer) bool {
			return p.TheirPieces().ContainPiece(pieceNo)
		})
	}
}

func filter(peers []peer.Peer, filterFunc func(peer.Peer) bool) []peer.Peer {
	var res []peer.Peer
	for _, p := range peers {
		if filterFunc(p) {
			res = append(res, p)
		}
	}
	return res

}

type PeerFilter func([]peer.Peer) []peer.Peer

func FilterPool(src []peer.Peer, filters ...PeerFilter) []peer.Peer {
	dst := src
	for _, f := range filters {
		dst = f(dst)
	}
	return dst
}
