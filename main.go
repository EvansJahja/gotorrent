package main

import (
	"fmt"
	"sync"

	"github.com/rapidloop/skv"

	"example.com/gotorrent/lib/platform/gcache"
	"example.com/gotorrent/lib/platform/udptracker"

	peerAdapter "example.com/gotorrent/lib/core/adapter/peer"
	"example.com/gotorrent/lib/platform/peer"

	"example.com/gotorrent/lib/platform/mem"
	"example.com/gotorrent/lib/platform/torrentdir"

	"example.com/gotorrent/lib/core/dag"
	"example.com/gotorrent/lib/core/dagrunner"
)

func main() {

	location := "/home/evans/torrent/test/"
	magnetURI := "***REMOVED***"

	skvStore, err := skv.Open(location + ".skv.db")
	if err != nil {
		panic(err)
	}

	/*
		v := dag.DownloadTorrent{
			Globals: dag.Globals{
				MagnetURI:  magnetURI,
				TargetPath: location,
				LocalRepo:  dummy.LocalRepo(),
			},
		}
	*/
	v := dag.DownloadStuffs{
		Globals: dag.Globals{
			MagnetURI: magnetURI,
			//TargetPath: location, // Maybe not use?

			LocalRepo: torrentdir.FileRepo{
				BasePath: location,
			},

			TempMetadata: mem.MemValueCache{
				SyncMap: &sync.Map{},
			},
			PersistentMetadata: skvStore,
			PeerFactory:        peerAdapter.PeerFactoryFn(peer.New),
			PeerList:           udptracker.New(),
			Cache:              gcache.NewCache(),
		},
	}
	dagrunner.Walk(v)
	fmt.Println("Program done")

	//svc := service.NewTorrent(magnetURI, location)
	/*
		svc := service.TorrentImpl{
			MagnetURI:  magnetURI,
			TargetPath: location,
			//PeerList:   dummy.PeerList(),
			PeerList: udptracker,
		}
		fmt.Println("Starting")
		<-svc.Start()
		fmt.Println("Done")
	*/

	// t := domain.Torrent{}

	// p := persistence.New(location, &t)
	// p.SetUpdateChan(t.GetUpdatedChan())

	// f := files.Files{Torrent: &t}
	// f.CreateFiles()

	// r := runner.Runner{
	// 	Torrent: &t,
	// 	Files:   f,
	// }

	// //r.SetupTracker()
	// r.Start()

	//f.GetLocalPiece(240)
	//f.WritePieceToLocal(1, []byte{0x0, 0x0, 0x0, 0x0})
	//f.WritePieceToLocal(1, []byte{0xDE, 0xAD, 0xBE, 0xEF})
	//x := f.GetLocalPiece(1)
	//fmt.Printf("%x", x[0:4])
	//f.CheckFiles()

	//t.Print()
	//return

	//t.Start()

}
