package main

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"example.com/gotorrent/lib/core/bucketdownload"
	"example.com/gotorrent/lib/core/service/peerlist"
	"example.com/gotorrent/lib/core/service/peerpool"
	"example.com/gotorrent/lib/files"

	"example.com/gotorrent/lib/core/domain"
	"example.com/gotorrent/lib/platform/gcache"
	"example.com/gotorrent/lib/platform/peer"
	"example.com/gotorrent/lib/platform/realclock"
	"example.com/gotorrent/lib/platform/udptracker"
	"example.com/gotorrent/lib/platform/upnp"

	"github.com/rapidloop/skv"
)

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	fmt.Println("Hi")
	defer func() {
		fmt.Println("Bye")
	}()

	location := "/home/evans/torrent/test/"
	magnetStr := "***REMOVED***"

	u, _ := url.Parse(magnetStr)
	magnetURI := domain.Magnet{Url: u}
	infoHash := magnetURI.InfoHash()
	trackers := magnetURI.Trackers()
	_ = trackers

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

	////////////////////////////////

	udpPeerList := udptracker.UdpPeerList{
		InfoHash: infoHash,
		Trackers: trackers,
		Clock:    realclock.RealClock{},
	}
	udpPeerList.Start()
	defer udpPeerList.Stop()

	hostList := peerlist.Impl{
		PersistentMetadata: skvStore,
		PeerList:           udpPeerList,
		Cache:              gcache.NewCache(),
	}
	_ = hostList

	var ourPieces domain.PieceList

	if err := skvStore.Get("pieces", &ourPieces); err != nil {
		ourPieces = domain.NewPieceList(torrentMeta.PiecesCount())
		skvStore.Put("pieces", ourPieces)
	}

	if checkedPieces, hasChanges := f.CheckPieces(ourPieces); hasChanges {
		fmt.Printf("has changes\n")
		ourPieces = checkedPieces
	}

	ourPiecesFn := func() domain.PieceList {
		return ourPieces
	}

	peerFactory := peer.PeerFactory{
		InfoHash:       infoHash,
		OurPieceListFn: ourPiecesFn,
	}

	peerPool := peerpool.Factory{
		PeerFactory: peer.PeerFactory{
			InfoHash:       infoHash,
			OurPieceListFn: ourPiecesFn,
		},
	}.New()

	/*
		hosts, err := hostList.GetHosts()
		if err != nil {
			fmt.Println(err.Error())
			return

		}
	*/
	//peerPool.AddHosts(hosts...)

	local := domain.Host{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 6882,
	}
	peerPool.AddHosts(local)

	peerPool.Start()

	newPeersChan, listenPort, err := peerFactory.Serve(infoHash)
	if err != nil {
		panic(err)
	}

	portExposer := upnp.New(uint16(listenPort))
	portExposer.Start()
	defer portExposer.Stop()

	<-quit
	return

	go func() {
		for newPeer := range newPeersChan {
			peerPool.AddPeer(newPeer)
		}
	}()

	go func() {
		for req := range peerPool.PieceRequests() {
			buf := f.GetLocalPiece(req.PieceNo)
			buf = buf[req.Begin : req.Begin+req.Length] // TODO: Refactor
			req.Response <- buf
		}
	}()

	fmt.Printf("Start asking for pieces\n")

	var wg sync.WaitGroup
	var pieceNo uint32
	conPieces := make(chan struct{}, 1)
	for pieceNo = 0; pieceNo <= 20; pieceNo++ {
		if ourPieces.ContainPiece(pieceNo) {
			continue
		}
		conPieces <- struct{}{}
		wg.Add(1)
		go func(pieceNo uint32) {
		RetryPiece:
			fmt.Printf("Start piece %d\n", pieceNo)

			fileWriteSeekerGen :=
				func() io.WriteSeeker {
					return f.WriteSeeker(int(pieceNo))
				}

			poolReaderGen := func() io.ReadSeekCloser {
				return peerPool.NewPeerPoolReader(pieceNo, f.Torrent.PieceLength, f.Torrent.PiecesCount(), f.Torrent.TorrentLength())
			}

			bd := bucketdownload.New(poolReaderGen, fileWriteSeekerGen, 1<<14, f.Torrent.PieceLength, 5)
			bd.Start()
			wg.Done()

			if ok := f.VerifyLocalPiece(pieceNo); !ok {
				fmt.Printf("piece %d corrupt, retrying...\n", pieceNo)
				time.Sleep(5 * time.Second)
				goto RetryPiece

			}
			ourPieces.SetPiece(pieceNo)
			if err := skvStore.Put("pieces", ourPieces); err != nil {
				fmt.Printf("error store pieces %s\n", err.Error())
			}

			<-conPieces
			fmt.Printf("Done piece %d\n", pieceNo)
		}(pieceNo)
	}
	wg.Wait()

	fmt.Printf("Closing app soon \n")
	time.Sleep(3 * time.Second)

}
