package main

import (
	"fmt"
	"net/url"

	"example.com/gotorrent/lib/core/domain"
	"example.com/gotorrent/lib/core/service/peerpool"
	"example.com/gotorrent/lib/files"
	"example.com/gotorrent/lib/platform/peer"
	"github.com/rapidloop/skv"
)

func main() {
	location := "/home/evans/torrent/test/"
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
	//f.CreateFiles()

	_ = f
	ourPieces := domain.NewPieceList(torrentMeta.PiecesCount())
	ourPiecesFn := func() domain.PieceList {
		return ourPieces
	}
	ourPieces.SetPiece(0)

	peerFactory := peer.PeerFactory{
		InfoHash:       infoHash,
		OurPieceListFn: ourPiecesFn,
	}

	newPeersChan, err := peerFactory.Serve(infoHash)
	if err != nil {
		panic(err)
	}

	peerPool := peerpool.Factory{
		PeerFactory: peerFactory,
	}.New()
	peerPool.Start()

	go func() {
		for req := range peerPool.PieceRequests() {
			buf := f.GetLocalPiece(int(req.PieceNo))
			buf = buf[req.Begin : req.Begin+req.Length] // This can't be good
			req.Response <- buf
		}
	}()

	go func() {
		for newPeer := range newPeersChan {
			peerPool.AddPeer(newPeer)
		}
	}()

	select {}

}
