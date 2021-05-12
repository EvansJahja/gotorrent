package peerpool

import (
	"fmt"
	"io"
	"sync"
	"time"

	"example.com/gotorrent/lib/core/adapter/peer"
	"example.com/gotorrent/lib/core/domain"
)

type Impl struct {
	PeerFactory    peer.PeerFactory
	connectedHosts []domain.Host
	newHosts       []domain.Host
	connectedPeers []peer.Peer
}

func (impl *Impl) AddHosts(newHosts ...domain.Host) {
	for _, newHost := range newHosts {
		for _, existingHost := range impl.newHosts {
			if newHost.Equal(existingHost) {
				return
			}
		}
		impl.newHosts = append(impl.newHosts, newHost)
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
		if len(impl.newHosts) == 0 {
			continue
		}

		var m sync.Mutex
		var wg sync.WaitGroup
		//var connectedPeers []peer.Peer

		for i, newHost := range impl.newHosts {
			wg.Add(1)
			go func(idx int, host domain.Host) {
				peer := impl.PeerFactory.New(host)
				impl.setupEventHandler(peer)
				err := peer.Connect()
				if err == nil {
					m.Lock()
					fmt.Printf("Connected to %s", string(peer.GetPeerID()))
					impl.connectedPeers = append(impl.connectedPeers, peer)
					m.Unlock()
				} else {
					m.Lock()
					m.Unlock()
				}
				wg.Done()
			}(i, newHost)
		}
		wg.Wait()
		impl.newHosts = []domain.Host{}

		fmt.Printf("Connected peers: %+v\n", impl.connectedPeers)

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
}
