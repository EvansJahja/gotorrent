package peerpool

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	"example.com/gotorrent/lib/core/adapter/peer"
	"example.com/gotorrent/lib/core/domain"
	"example.com/gotorrent/lib/logger"
	"go.uber.org/ratelimit"
	"go.uber.org/zap"
)

var l_peerpool = logger.Named("peerpool")

type PeerPool interface {
	AddPeer(newPeers ...peer.Peer)
	AddHosts(newHosts ...domain.Host)
	NewPeerPoolReader(pieceNo uint32, pieceLength int, pieceCount int, torrentLength int) io.ReadSeekCloser
	Start()

	FindNextPiece(have domain.PieceList, pieceCount int) (uint32, error)
	PieceRequests() <-chan peer.PieceRequest
	Peers(filters ...PeerFilter) []peer.Peer

	SetInterestedIn(pieceNo uint32)
	// Call this when piece is done, to send Have to peers
	TellPieceCompleted(pieceNo uint32)

	GetNetworkStats() (downloadRate, uploadRate float32, downloadBytes, uploadBytes uint64)
}
type Factory struct {
	PeerFactory interface {
		New(host domain.Host) peer.Peer
	}
}

func (b Factory) New() PeerPool {
	return &peerPoolImpl{
		Factory:      b,
		newHosts:     make(chan domain.Host),
		pieceRequest: make(chan peer.PieceRequest),
	}
}

type peerPoolImpl struct {
	Factory

	connectedHosts []domain.Host
	connectedPeers []peer.Peer
	newHosts       chan domain.Host
	pieceRequest   chan peer.PieceRequest
}

func (impl *peerPoolImpl) PieceRequests() <-chan peer.PieceRequest {
	return impl.pieceRequest
}

func (impl *peerPoolImpl) AddPeer(newPeers ...peer.Peer) {
	for _, p := range newPeers {
		impl.runPeer(p)
	}
}

func (impl *peerPoolImpl) AddHosts(newHosts ...domain.Host) {
	for _, newHost := range newHosts {
		for _, existingHost := range impl.connectedHosts {
			if newHost.Equal(existingHost) {
				return
			}
		}
		go func(newHost domain.Host) {
			impl.newHosts <- newHost
		}(newHost)
		//impl.newHosts = append(impl.newHosts, newHost)
	}
}

func (impl *peerPoolImpl) Start() {
	fmt.Printf("starting peerpool\n")
	go impl.run()
}

func (impl *peerPoolImpl) NewPeerPoolReader(pieceNo uint32, pieceLength int, pieceCount int, torrentLength int) io.ReadSeekCloser {
	return &poolReaderImpl{impl: impl,
		pieceNo:       pieceNo,
		pieceLength:   pieceLength,
		pieceCount:    pieceCount,
		torrentLength: torrentLength,
	}

}

func (impl *peerPoolImpl) run() {
	ticker := time.NewTicker(1 * time.Second)
	//Ticker:
	for range ticker.C {
		//fmt.Printf("processing %d new hosts: \n", len(impl.newHosts))
		if len(impl.connectedPeers) >= 5 {
			continue
		}
		/*
			if len(impl.newHosts) == 0 {
				continue
			}
		*/

		var wg sync.WaitGroup
		//var connectedPeers []peer.Peer

		for newHost := range impl.newHosts {
			wg.Add(1)
			go func(host domain.Host) {
				peer := impl.PeerFactory.New(host)
				impl.runPeer(peer)
				wg.Done()
			}(newHost)
		}
		wg.Wait()

		fmt.Printf("Connected peers: %+v\n", impl.connectedPeers)

	}
}

func (impl *peerPoolImpl) runPeer(p peer.Peer) {
	l_peerpool.Sugar().Debugw("add peer to pool", "peer", p.GetID())
	impl.setupEventHandler(p)
	if p.GetState().Connected {
		// Fast track
		impl.connectedPeers = append(impl.connectedPeers, p)
		return

	}
RetryConnect:
	err := p.Connect()

	if err == nil {
		impl.connectedPeers = append(impl.connectedPeers, p)
	} else {
		time.Sleep(5 * time.Second)
		goto RetryConnect
	}

}

func (impl *peerPoolImpl) setupEventHandler(p peer.Peer) {

	/*
		p.OnPiecesUpdatedChanged(func() {
			p.Interested()

		})
	*/
	/*
		p.OnChokedChanged(func(isChoked bool) {
			if isChoked {
				fmt.Printf("%s is choked\n", p.GetPeerID())
			} else {
				fmt.Printf("%s is unchoked\n", p.GetPeerID())
			}

		})
	*/
	go func() {
		rl := ratelimit.New(10)
		for req := range p.PieceRequests() {
			rl.Take()
			l_peerpool.Sugar().Debugw("got request", "pieceno", req.PieceNo, "begin", req.Begin)
			impl.pieceRequest <- req
		}
		panic("chan Closed")

	}()
}
func (impl *peerPoolImpl) FindNextPiece(havePiece domain.PieceList, pieceCount int) (uint32, error) {

	wantPieces := make([]uint32, 0, pieceCount)

	for i := uint32(0); i < uint32(pieceCount); i++ {
		if !havePiece.ContainPiece(i) {
			wantPieces = append(wantPieces, i)
		}
	}
	if len(wantPieces) == 0 {
		return 0, errors.New("no want pieces")
	}

	l_peerpool.Debug("finding next piece")

Retry:

	for len(impl.connectedPeers) == 0 {
		l_peerpool.Debug("waiting for any peer to connect")
		time.Sleep(1 * time.Second)
	}
	l_peerpool.Debug("got peers", zap.Int("connectedPeers", len(impl.connectedPeers)))

	maxV := 0
	minV := 0
	pieceCounts := make(map[uint32]int)
	for _, p := range impl.connectedPeers {
		theirPiece := p.TheirPieces()
		for _, i := range wantPieces {
			if theirPiece.ContainPiece(i) {
				pieceCounts[i] += 1
			}
		}
	}

	for _, v := range pieceCounts {
		if maxV < v {
			maxV = v
		}
		if minV > v || minV == 0 {
			minV = v
		}
	}
	if len(pieceCounts) == 0 {
		l_peerpool.Debug("piece count is 0, retrying")
		time.Sleep(5 * time.Second)
		goto Retry

	}

	// minV should NEVER be 0, so this is a precaution
	if minV == 0 {
		panic("minV should not be 0")
	}

	if len(pieceCounts) == 0 {
		panic("should be unreachable ")
	}

	peerHasPieces := make([]uint32, 0, pieceCount)
	rarePieces := make([]uint32, 0, pieceCount)
	for k, v := range pieceCounts {
		if v == minV {
			rarePieces = append(rarePieces, k)
		}
		peerHasPieces = append(peerHasPieces, k)
	}
	if len(pieceCounts) == 0 {
		panic("should be unreachable ")
	}

	if maxV == 1 || minV == 0 {
		pieceIdx := peerHasPieces[rand.Int()%len(peerHasPieces)]
		l_peerpool.Debug("choosing next piece randomly", zap.Uint32("pieceIdx", pieceIdx))
		return pieceIdx, nil
	}

	if len(pieceCounts) == 0 {
		panic("should be unreachable ")
	}

	pieceIdx := rarePieces[rand.Int()%len(rarePieces)]
	l_peerpool.Debug("choosing next piece with rare pieces", zap.Uint32("pieceIdx", pieceIdx), zap.Uint32s("rare pieces", rarePieces), zap.Int("minV", minV))
	return pieceIdx, nil

}

func (impl *peerPoolImpl) TellPieceCompleted(pieceNo uint32) {
	for _, p := range impl.connectedPeers {
		p.TellPieceCompleted(pieceNo)
	}

}

func (impl *peerPoolImpl) GetNetworkStats() (downloadRate, uploadRate float32, downloadBytes, uploadBytes uint64) {
	for _, p := range impl.connectedPeers {
		downloadRate += p.GetDownloadRate()
		uploadRate += p.GetUploadRate()
		downloadBytes += uint64(p.GetDownloadBytes())
		uploadBytes += uint64(p.GetUploadBytes())
	}

	return

}

func (impl *peerPoolImpl) Peers(filters ...PeerFilter) []peer.Peer {
	return FilterPool(impl.connectedPeers, filters...)
}

func (impl *peerPoolImpl) SetInterestedIn(pieceNo uint32) {
	// 1. find all peers that "we are interested", but does not have pieceNo, tell not interested
	// 2. find all peers that "we are not interested", but have piece no, tell interested
	tellNotInterestedPeers := impl.Peers(FilterConnected, FilterWeAreInterested, FilterNotHavePiece(pieceNo))

	for _, p := range tellNotInterestedPeers {
		p.Uninterested()
	}
	tellInterestedPeers := impl.Peers(FilterConnected, FilterWeAreNotInterested, FilterHasPiece(pieceNo))
	for _, p := range tellInterestedPeers {
		p.Interested()
	}

}
