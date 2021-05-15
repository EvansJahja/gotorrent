package peerpool

import (
	"fmt"
	"io"
	"sync"
	"time"

	"example.com/gotorrent/lib/core/adapter/peer"
	"example.com/gotorrent/lib/core/domain"
	"example.com/gotorrent/lib/logger"
)

var l_peerpool = logger.Named("peerpool")

type PeerPool interface {
	AddPeer(newPeers ...peer.Peer)
	AddHosts(newHosts ...domain.Host)
	NewPeerPoolReader(pieceNo uint32, pieceLength int, pieceCount int, torrentLength int) io.ReadSeekCloser
	Start()

	PieceRequests() <-chan peer.PieceRequest
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
	impl.setupEventHandler(p)
	if p.GetState().Connected {
		// Fast track
		impl.connectedPeers = append(impl.connectedPeers, p)
		return

	}
RetryConnect:
	err := p.Connect()

	if err == nil {
		fmt.Printf("W Connected to %s\n", string(p.GetPeerID()))
		impl.connectedPeers = append(impl.connectedPeers, p)
	} else {
		time.Sleep(5 * time.Second)
		goto RetryConnect
	}

}

func (impl *peerPoolImpl) setupEventHandler(p peer.Peer) {

	p.OnPiecesUpdatedChanged(func() {
		p.Interested()

	})
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
		for req := range p.PieceRequests() {
			l_peerpool.Sugar().Debugw("got request", "pieceno", req.PieceNo, "begin", req.Begin)
			impl.pieceRequest <- req
		}
		panic("chan Closed")

	}()
}
