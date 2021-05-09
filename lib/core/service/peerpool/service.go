package peerpool

import (
	"fmt"
	"sync"
	"time"

	"example.com/gotorrent/lib/core/adapter/peer"
	"example.com/gotorrent/lib/core/domain"
)

type Service interface {
	AddHosts(newHosts ...domain.Host)
}
type Impl struct {
	PeerFactory    peer.PeerFactory
	connectedHosts []domain.Host
	newHosts       []domain.Host
	connectedPeers []peer.Peer
}

func (impl *Impl) AddHosts(newHosts ...domain.Host) {
	impl.newHosts = append(impl.newHosts, newHosts...)
}

func (impl *Impl) Start() {
	fmt.Printf("starting peerpool\n")
	go impl.run()
}

func (impl *Impl) run() {
	ticker := time.NewTicker(5 * time.Second)
	//Ticker:
	for range ticker.C {
		fmt.Printf("processing %d new hosts: \n", len(impl.newHosts))
		if len(impl.connectedPeers) >= 5 {
			continue
		}
		if len(impl.newHosts) == 0 {
			continue
		}

		var m sync.Mutex
		var wg sync.WaitGroup
		var connectedPeers []peer.Peer

		for i, newHost := range impl.newHosts {
			wg.Add(1)
			go func(idx int, host domain.Host) {
				peer := impl.PeerFactory.New(host)
				impl.setupEventHandler(peer)
				err := peer.Connect()
				if err == nil {
					m.Lock()
					fmt.Printf("Connected to %s", string(peer.GetPeerID()))
					connectedPeers = append(connectedPeers, peer)
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

		fmt.Printf("Connected peers: %+v\n", connectedPeers)

	}
}

func (impl *Impl) setupEventHandler(p peer.Peer) {

	p.OnPiecesUpdatedChanged(func() {
		p.Unchoke()
		p.Interested()
		p.RequestPiece(0, 0, 8)

	})
	p.OnChokedChanged(func(isChoked bool) {
		if isChoked {
			fmt.Printf("%s is choked\n", p.GetPeerID())
		} else {
			fmt.Printf("%s is unchoked\n", p.GetPeerID())
		}

	})
}
