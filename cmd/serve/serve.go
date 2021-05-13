package main

import (
	"fmt"
	"net/url"

	peerAdapter "example.com/gotorrent/lib/core/adapter/peer"
	"example.com/gotorrent/lib/core/domain"
	"example.com/gotorrent/lib/core/service/peerpool"
	"example.com/gotorrent/lib/files"
	"example.com/gotorrent/lib/platform/peer"
	"github.com/rapidloop/skv"
)

func main() {
	location := "/home/evans/torrent/test2/"
	magnetStr := "***REMOVED***"

	u, _ := url.Parse(magnetStr)
	magnetURI := domain.Magnet{Url: u}
	infoHash := magnetURI.InfoHash()

	skvStore, err := skv.Open(location + ".skv.db")
	if err != nil {
		panic(err)
	}
	defer skvStore.Close()

	var metadata domain.Metadata
	err = skvStore.Get("metadata", &metadata)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Received %x\n", metadata.InfoHash())
	fmt.Printf("Expected %x\n", infoHash)

	torrentMeta := metadata.MustParse()
	f := files.Files{Torrent: torrentMeta, BasePath: location}

	_ = f
	newPeersChan, err := peer.Serve(infoHash)
	if err != nil {
		panic(err)
	}

	peerPool := peerpool.Factory{
		PeerFactory: peerAdapter.NewPeerFactory(infoHash, peer.New),
	}.New()
	peerPool.Start()

	go func() {
		for newPeer := range newPeersChan {
			peerPool.AddPeer(newPeer)
		}
	}()

	select {}

}
