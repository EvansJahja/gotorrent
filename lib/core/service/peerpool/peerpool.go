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

type Impl struct {
	PeerFactory    peer.PeerFactory
	connectedHosts []domain.Host
	NewHosts       chan domain.Host
	//newHosts       []domain.Host
	connectedPeers []peer.Peer
}

func (impl *Impl) AddPeer(newPeers ...peer.Peer) {
	for _, p := range newPeers {
		impl.runPeer(p)
	}
}
func (impl *Impl) AddHosts(newHosts ...domain.Host) {
	for _, newHost := range newHosts {
		for _, existingHost := range impl.connectedHosts {
			if newHost.Equal(existingHost) {
				return
			}
		}
		go func(newHost domain.Host) {
			impl.NewHosts <- newHost
		}(newHost)
		//impl.newHosts = append(impl.newHosts, newHost)
	}
}

func (impl *Impl) Start() {
	fmt.Printf("starting peerpool\n")
	go impl.run()
}

func (impl *Impl) NewPeerPoolReader(pieceNo uint32, pieceLength int, pieceCount int, torrentLength int) io.ReadSeekCloser {
	return &poolReaderImpl{impl: impl,
		pieceNo:       pieceNo,
		pieceLength:   pieceLength,
		pieceCount:    pieceCount,
		torrentLength: torrentLength,
	}

}

func (impl *Impl) run() {
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

		for newHost := range impl.NewHosts {
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

func (impl *Impl) runPeer(p peer.Peer) {
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

func (impl *Impl) setupEventHandler(p peer.Peer) {

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
