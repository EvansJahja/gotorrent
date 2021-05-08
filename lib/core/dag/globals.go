package dag

import (
	"example.com/gotorrent/lib/core/adapter/cache"
	"example.com/gotorrent/lib/core/adapter/localrepo"
	"example.com/gotorrent/lib/core/adapter/peer"
	"example.com/gotorrent/lib/core/adapter/peerlist"
	"example.com/gotorrent/lib/core/adapter/persistentmetadata"
	"example.com/gotorrent/lib/core/adapter/temprepo"
)

type Globals struct {
	MagnetURI string
	//TargetPath string
	// Runner     func(n Node)
	Peers []peer.Peer

	PeerFactory        peer.PeerFactory
	LocalRepo          localrepo.LocalRepo
	TempMetadata       temprepo.TempMetadata
	PersistentMetadata persistentmetadata.PersistentMetadata
	PeerList           peerlist.PeerRepo
	Cache              cache.Cache
}
