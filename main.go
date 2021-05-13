package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"net/url"
	"sync"
	"time"

	"example.com/gotorrent/lib/core/bucketdownload"
	"example.com/gotorrent/lib/core/service/peerpool"
	"example.com/gotorrent/lib/files"

	"example.com/gotorrent/lib/core/domain"
	"example.com/gotorrent/lib/platform/peer"

	"github.com/rapidloop/skv"
)

func Noise() io.ReadSeekCloser {
	return impl{}
}

type impl struct{}

func (im impl) Read(p []byte) (n int, err error) {
	return rand.Read(p)
}

func (im impl) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (im impl) Close() error {
	return nil
}

func main() {
	// DOWNLOADED: 0, 241
	location := "/home/evans/torrent/test2/"
	magnetStr := "***REMOVED***"

	u, _ := url.Parse(magnetStr)
	magnetURI := domain.Magnet{Url: u}
	infoHash := magnetURI.InfoHash()
	trackers := magnetURI.Trackers()
	_ = trackers
	//v := u.Query()
	//trackers := v["tr"]

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
	// WARNING
	// f.CreateFiles()
	// WARNING

	//f.CheckFiles()

	////////////////////////////////

	/*
		hostList := peerlist.Impl{
			PersistentMetadata: skvStore,
			PeerList: udptracker.UdpPeerList{
				InfoHash: infoHash,
				Trackers: trackers,
			},
			Cache: gcache.NewCache(),
		}
	*/
	ourPieces := domain.NewPieceList(torrentMeta.PiecesCount())
	//ourPieces.SetPiece(0)

	peerPool := peerpool.Factory{
		PeerFactory: peer.PeerFactory{
			InfoHash:     infoHash,
			OurPieceList: ourPieces,
		},
	}.New()

	/*
		hosts, err := hostList.GetHosts()
		if err != nil {
			fmt.Println(err.Error())
			return

		}
	*/
	//_ = hosts

	//_ = hosts
	//_ = peerPool
	peerPool.Start()
	//peerPool.AddHosts(hosts...)
	local := domain.Host{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 6881,
	}
	peerPool.AddHosts(local)

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

	//fmt.Printf("piece count %d\n", len(f.Torrent.Pieces)/20) // 242

	//fmt.Printf("Piece length:  %d\n", torrentMeta.PieceLength)

	/*

		var x int
		ptr := passthroughreader.NewPassthrough(r, func(n int) {
			x += n
			prog := float32(100.0*x) / 16777216
			fmt.Printf("Progress: %f%%\n", prog)
		})
	*/

	//seekStart := int64(668725 + 1771504)
	//b := make([]byte, 100)
	////io.CopyBuffer(io.Discard, ptr, b)

	// We already done piece 0
	//os.Exit(1)
	fmt.Printf("Start asking for pieces\n")

	var wg sync.WaitGroup
	var pieceNo uint32
	conPieces := make(chan struct{}, 1)
	for pieceNo = 0; pieceNo <= 0; pieceNo++ {
		conPieces <- struct{}{}
		wg.Add(1)
		go func(pieceNo uint32) {
			fmt.Printf("Start piece %d\n", pieceNo)

			fileWriteSeekerGen :=
				func() io.WriteSeeker {
					return f.WriteSeeker(int(pieceNo))
				}

			poolReaderGen := func() io.ReadSeekCloser {

				//return Noise()
				return peerPool.NewPeerPoolReader(pieceNo, f.Torrent.PieceLength, f.Torrent.PiecesCount(), f.Torrent.TorrentLength())
			}

			bd := bucketdownload.New(poolReaderGen, fileWriteSeekerGen, 1<<14, f.Torrent.PieceLength, 10)
			bd.Start()
			wg.Done()
			<-conPieces
			fmt.Printf("Done piece %d\n", pieceNo)
		}(pieceNo)
	}
	wg.Wait()

	/*
		//r1 := peerPool.NewPeerPoolReader(0, 16777216)
		r1 := poolReaderGen()
		seekStart := int64(0)
		r1.Seek(seekStart, io.SeekStart)
		r1Lim := io.LimitReader(r1, 3322349)

		bufRead1 := bufio.NewReader(r1Lim)
		//_ = fileWriteSeekerGen
		//f.WritePieceToLocal(0, bufRead1, seekStart)
		writeSeeker := fileWriteSeekerGen()
		writeSeeker.Seek(seekStart, io.SeekStart)
		n, err := io.Copy(writeSeeker, bufRead1)
		if err == nil {
			// EOF
			return
		}
		if err != nil {
			fmt.Print("o")
		}
		if n != 0 {
			fmt.Print("o")
		}
	*/

	//f.WritePieceToLocal(0, bufRead1, seekStart)
	/*
		var wg sync.WaitGroup

		wg.Add(2)
		go func() {
			r1 := peerPool.NewPeerPoolReader(0, 16777216)
			seekStart := int64(10132517)
			r1.Seek(seekStart, io.SeekStart)

			r1Lim := io.LimitReader(r1, 3322349)

			bufRead1 := bufio.NewReader(r1Lim)
			f.WritePieceToLocal(0, bufRead1, seekStart)
			wg.Done()
		}()
		go func() {
			r2 := peerPool.NewPeerPoolReader(0, 16777216)
			seekStart := int64(10132517 + 3322349)
			r2.Seek(seekStart, io.SeekStart)

			r2Lim := io.LimitReader(r2, 13454867)

			bufRead1 := bufio.NewReader(r2Lim)
			f.WritePieceToLocal(0, bufRead1, seekStart)
			wg.Done()
		}()
		wg.Wait()
	*/

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
	time.Sleep(3 * time.Second)

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
