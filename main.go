package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/gotorrent/lib/core/bucketdownload"
	"example.com/gotorrent/lib/core/service/peerlist"
	"example.com/gotorrent/lib/core/service/peerpool"
	"example.com/gotorrent/lib/files"
	"example.com/gotorrent/lib/logger"
	"example.com/gotorrent/lib/transport/echohttp"
	"go.uber.org/zap"

	"example.com/gotorrent/lib/core/domain"
	"example.com/gotorrent/lib/platform/gcache"
	"example.com/gotorrent/lib/platform/peer"
	"example.com/gotorrent/lib/platform/udptracker"
	"example.com/gotorrent/lib/platform/upnp"

	"github.com/rapidloop/skv"
)

func main() {

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	l := logger.Named("main")
	l.Info("Hi")

	defer func() {
		l.Info("Bye")
	}()

	location := "/home/evans/torrent/test/"
	magnetStr := "***REMOVED***"

	l.Info("download location: " + location)
	u, _ := url.Parse(magnetStr)
	magnetURI := domain.Magnet{Url: u}
	infoHash := magnetURI.InfoHash()
	trackers := magnetURI.Trackers()

	skvStore, err := skv.Open(location + ".skv.db")
	{
		if err != nil {
			panic(err)
		}
		defer skvStore.Close()
	}

	var metadata domain.Metadata
	{
		err = skvStore.Get("metadata", &metadata)
		if err != nil {
			panic(err)
		}
		l.Sugar().Debugw("check our cached infohash", "expected", hex.EncodeToString(infoHash), "actual", hex.EncodeToString(metadata.InfoHash()))
	}

	torrentMeta := metadata.MustParse()

	f := files.Files{Torrent: torrentMeta, BasePath: location}
	var ourPieces domain.PieceList
	{
		if err := skvStore.Get("pieces", &ourPieces); err != nil {
			ourPieces = domain.NewPieceList(torrentMeta.PiecesCount())
			skvStore.Put("pieces", ourPieces)
		}
		if checkedPieces, hasChanges := f.CheckPieces(ourPieces); hasChanges {
			fmt.Printf("has changes\n")
			ourPieces = checkedPieces
		}
	}

	ourPiecesFn := func() domain.PieceList {
		return ourPieces
	}

	peerFactory := peer.PeerFactory{
		InfoHash:       infoHash,
		OurPieceListFn: ourPiecesFn,
	}
	newPeersChan, listenPort, err := peerFactory.Serve(infoHash)
	if err != nil {
		panic(err)
	}
	portExposer := upnp.New(uint16(listenPort))
	portExposer.Start()
	defer portExposer.Stop()

	trackerInfo := udptracker.TrackerInfo{
		Uploaded:   0,
		Downloaded: 0,
		Left:       0,
		Port:       portExposer.Port(),
	}

	udpPeerList := &udptracker.UdpPeerList{
		InfoHash: infoHash,
		Trackers: trackers,
	}
	udpPeerList.SetInfo(trackerInfo)
	udpPeerList.Start()

	hostList := peerlist.Impl{
		PersistentMetadata: skvStore,
		PeerList:           udpPeerList,
		Cache:              gcache.NewCache(),
	}

	peerPool := peerpool.Factory{
		PeerFactory: peer.PeerFactory{
			InfoHash:       infoHash,
			OurPieceListFn: ourPiecesFn,
		},
	}.New()

	go func() {
		for newPeer := range newPeersChan {
			peerPool.AddPeer(newPeer)
		}
	}()

	hosts, err := hostList.GetHosts()

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	l.Sugar().Debugw("host list", "hosts", hosts)

	peerPool.AddHosts(hosts...)

	/*
		local := domain.Host{
			IP:   net.IPv4(127, 0, 0, 1),
			Port: 6882,
		}
		peerPool.AddHosts(local)
	*/

	peerPool.Start()

	go func() {
		for req := range peerPool.PieceRequests() {
			buf := f.GetLocalPiece(req.PieceNo)
			buf = buf[req.Begin : req.Begin+req.Length] // TODO: Refactor
			req.Response <- buf
		}
	}()

	httpServe := echohttp.HTTPServe{
		PeerPool: peerPool,
	}
	httpServe.Start()

	l.Info("Start asking for pieces")

	go func() {
		for {
			l := logger.Named("stats")
			downloadRate, uploadRate, downloadBytes, uploadBytes := peerPool.GetNetworkStats()
			l.Sugar().Infof("\nDown\t%f\nUp\t%f\nDownTot\t%f\nUpTot\t%f\n", downloadRate, uploadRate, float64(downloadBytes)/1e6, float64(uploadBytes)/1e6)

			time.Sleep(5 * time.Second)
		}
	}()

	for {
		pieceNo, err := peerPool.FindNextPiece(ourPieces, f.Torrent.PiecesCount())
		if err != nil {
			// No more piece
			// Todo: check error
			break
		}
		if ourPieces.ContainPiece(pieceNo) {
			panic("should not happen")
			continue
		}
		// TODO  tell interested and not interested based on our pieceNo
	RetryPiece:

		l.Sugar().Infof("Start downloading piece %d", pieceNo)

		fileWriteSeekerGen :=
			func() io.WriteSeeker {
				return f.WriteSeeker(int(pieceNo))
			}

		poolReaderGen := func() io.ReadSeekCloser {
			return peerPool.NewPeerPoolReader(pieceNo, f.Torrent.PieceLength, f.Torrent.PiecesCount(), f.Torrent.TorrentLength())
		}

		bd := bucketdownload.New(poolReaderGen, fileWriteSeekerGen, 1<<14, f.Torrent.PieceLength, 5)
		bd.Start()

		if ok := f.VerifyLocalPiece(pieceNo); !ok {
			l.Sugar().Errorf("piece %d corrupt, retrying...", pieceNo)
			time.Sleep(5 * time.Second)
			goto RetryPiece

		}

		ourPieces.SetPiece(pieceNo)
		peerPool.TellPieceCompleted(pieceNo)

		if err := skvStore.Put("pieces", ourPieces); err != nil {
			l.Error("error store piece", zap.Error(err))
		}

		l.Sugar().Infof("Done download and verify piece %d", pieceNo)

		// TODO
		// 1. Update our bitfields -> ourPiecesFn
		// 2. Send "Have" to peers -> peerPool.OnPieceCompleted
		// 3. Tell trackers

		// tell trackers
		_, _, downloadBytes, uploadBytes := peerPool.GetNetworkStats()

		trackerInfo := udptracker.TrackerInfo{
			Uploaded:   int(uploadBytes),
			Downloaded: int(downloadBytes),
			Left:       0,
			Port:       portExposer.Port(),
		}
		udpPeerList.SetInfo(trackerInfo)

	}

	<-quit
	return

}
