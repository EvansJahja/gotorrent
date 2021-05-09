package main

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"time"

	peerAdapter "example.com/gotorrent/lib/core/adapter/peer"
	"example.com/gotorrent/lib/core/service/peerlist"

	"example.com/gotorrent/lib/core/domain"
	"example.com/gotorrent/lib/platform/gcache"
	"example.com/gotorrent/lib/platform/peer"
	"example.com/gotorrent/lib/platform/udptracker"

	"github.com/rapidloop/skv"
)

func main() {
	location := "/home/evans/torrent/test/"
	magnetURI := "***REMOVED***"

	infoHashStr := "***REMOVED***"
	infoHash, _ := hex.DecodeString(infoHashStr)

	u, _ := url.Parse(magnetURI)
	v := u.Query()
	trackers := v["tr"]

	skvStore, err := skv.Open(location + ".skv.db")
	if err != nil {
		panic(err)
	}
	defer skvStore.Close()

	hostList := peerlist.Impl{
		PersistentMetadata: skvStore,
		PeerList: udptracker.UdpPeerList{
			InfoHash: infoHash,
			Trackers: trackers,
		},
		Cache: gcache.NewCache(),
	}
	var _ = hostList
	/*

		peerPool := peerpool.Impl{
			PeerFactory: peerAdapter.NewPeerFactory(infoHash, peer.New),
		}

		hosts, err := hostList.GetHosts()
		if err != nil {
			fmt.Println(err.Error())
			return

		}
	*/
	/*
		peerPool.Start()
		peerPool.AddHosts(hosts...)

	*/
	peerFactory := peerAdapter.NewPeerFactory(infoHash, peer.New)
	targetHost := domain.Host{
		IP:   []byte{99, 232, 180, 37},
		Port: 56555,
	}

	var metadata domain.Metadata
	err = skvStore.Get("metadata", &metadata)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Received %x\n", metadata.InfoHash())
	fmt.Printf("Expected %s\n", infoHashStr)

	torrentMeta := metadata.MustParse()

	fmt.Printf("%+v\n", torrentMeta)

	p := peerFactory.New(targetHost)
	p.OnChokedChanged(func(isChoked bool) {

	})
	p.OnPiecesUpdatedChanged(func() {
		//p.Unchoke()
		//p.Interested()
		/*
			metadata, err := p.GetMetadata()
			if err == nil {
				fmt.Printf("Received %x\n", metadata.InfoHash())
				fmt.Printf("Expected %s\n", infoHashStr)
				skvStore.Put("metadata", metadata)
				fmt.Print("done")

			}
		*/

	})
	p.Connect()

	<-time.After(10 * time.Second)

	/*
		v := dag.DownloadTorrent{
			Globals: dag.Globals{
				MagnetURI:  magnetURI,
				TargetPath: location,
				LocalRepo:  dummy.LocalRepo(),
			},
		}
	*/
	/*
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
	*/

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
