package peerpool

import (
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	"example.com/gotorrent/lib/core/adapter/peer"
	"example.com/gotorrent/lib/core/domain"
)

type PeerPool interface {
	AddPeer(newPeers ...peer.Peer)
	AddHosts(newHosts ...domain.Host)
	NewPeerPoolReader(pieceNo uint32, pieceLength int, pieceCount int, torrentLength int) io.ReadSeekCloser
	Start()

	PieceRequests() <-chan peer.PieceRequest
}
type Factory struct {
	PeerFactory peer.PeerFactory
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
	var m sync.Mutex
	impl.setupEventHandler(p)
	if p.GetState().Connected {
		// Fast track
		impl.connectedPeers = append(impl.connectedPeers, p)
		return

	}
	err := p.Connect()
	if err == nil {
		m.Lock()
		fmt.Printf("AAAA\n")
		fmt.Printf("W Connected to %s", string(p.GetPeerID()))
		impl.connectedPeers = append(impl.connectedPeers, p)
		m.Unlock()
	} else {
		m.Lock()
		m.Unlock()
	}

}

func (impl *peerPoolImpl) setupEventHandler(p peer.Peer) {

	p.OnPiecesUpdatedChanged(func() {
		p.Interested()

	})
	p.OnChokedChanged(func(isChoked bool) {
		if isChoked {
			fmt.Printf("%s is choked\n", p.GetPeerID())
		} else {
			fmt.Printf("%s is unchoked\n", p.GetPeerID())
		}

	})
	go func() {
		fmt.Println("waiting")
		for req := range p.PieceRequests() {
			fmt.Printf("got request #%d %d", req.PieceNo, req.Begin)
			resp := make([]byte, 0, req.Length)
			for i := 0; i < int(req.Length); i++ {
				b := byte(rand.Int())
				resp = append(resp, b)
			}
			req.Response <- resp
			fmt.Printf("responding with random #%d %d", req.PieceNo, req.Begin)
		}
		panic("chan Closed")

	}()
}
