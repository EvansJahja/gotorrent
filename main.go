package main

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"time"

	peerAdapter "example.com/gotorrent/lib/core/adapter/peer"
	"example.com/gotorrent/lib/core/service/peerlist"
	"example.com/gotorrent/lib/core/service/peerpool"
	"example.com/gotorrent/lib/files"
	"example.com/gotorrent/lib/passthroughreader"

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

	peerPool := peerpool.Impl{
		PeerFactory: peerAdapter.NewPeerFactory(infoHash, peer.New),
	}

	hosts, err := hostList.GetHosts()
	if err != nil {
		fmt.Println(err.Error())
		return

	}
	peerPool.Start()
	peerPool.AddHosts(hosts...)

	/*
		peerFactory := peerAdapter.NewPeerFactory(infoHash, peer.New)
		targetHost1 := domain.Host{
			IP:   []byte{99, 232, 180, 37},
			Port: 56555,
		}
		targetHost2 := domain.Host{
			IP:   []byte{41, 107, 76, 73},
			Port: 37746,
		}
	*/

	var metadata domain.Metadata
	err = skvStore.Get("metadata", &metadata)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Received %x\n", metadata.InfoHash())
	fmt.Printf("Expected %s\n", infoHashStr)

	torrentMeta := metadata.MustParse()

	fmt.Printf("Piece length:  %d\n", torrentMeta.PieceLength)

	r := peerPool.NewPeerPoolReader(0, 16777216)

	var x int
	ptr := passthroughreader.NewPassthrough(r, func(n int) {
		x += n
		prog := float32(100.0*x) / 16777216
		fmt.Printf("Progress: %f%%\n", prog)
	})
	_ = ptr
	//b := make([]byte, 9900)
	////io.CopyBuffer(io.Discard, ptr, b)

	f := files.Files{Torrent: &torrentMeta, BasePath: location}
	f.WritePieceToLocal(0, ptr)

	/*
		fmt.Println("1")
		io.CopyN(io.Discard, r, 1024)
		fmt.Println("2")
		io.CopyN(io.Discard, r, 1024)
		fmt.Println("3")
		io.CopyN(io.Discard, r, 1024)
		fmt.Println("4")
	*/
	//kio.Copy(io.Discard, r)

	fmt.Printf("Closing app soon \n")
	time.Sleep(10 * time.Second)

	/*
		p1 := peerFactory.New(targetHost1)
		p2 := peerFactory.New(targetHost2)

		p1.OnPiecesUpdatedChanged(func() {
			p1.Unchoke()
			p1.Interested()
		})
		p2.OnPiecesUpdatedChanged(func() {
			p2.Unchoke()
			p2.Interested()
		})

		p1.Connect()
		p2.Connect()

		pr1 := peerAdapter.NewPeerReader(p1, 0, 16777216)
		pr2 := peerAdapter.NewPeerReader(p2, 0, 16777216)

		r, w := io.Pipe()

		readers := []io.ReadSeeker{pr1, pr2}
		cur := int64(0)
		go func() {
		Loop:
			for {
				for _, peerR := range readers {
					peerR.Seek(cur, io.SeekStart)
					n, err := io.CopyN(w, peerR, 2048)
					if err == io.EOF {
						break Loop
					}
					cur += n
				}
			}
		}()

		bufbytes := make([]byte, 16777216)
		buf := bytes.NewBuffer(bufbytes)

		var x int
		ptr := passthroughreader.NewPassthrough(r, func(n int) {
			x += n
			prog := float32(100.0*x) / 16777216
			fmt.Printf("Progress: %f%%\n", prog)
		})

		io.Copy(buf, ptr)

		//fmt.Printf("%d %s", n, err.Error())

		fmt.Printf("%s\n", buf)
		fmt.Printf("ok")

		<-time.After(60 * time.Second)
	*/

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
