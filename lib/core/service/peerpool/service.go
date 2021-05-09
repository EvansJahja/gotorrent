package peerpool

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"example.com/gotorrent/lib/core/adapter/peer"
	"example.com/gotorrent/lib/core/domain"
)

type Service interface {
	AddHosts(newHosts ...domain.Host)
	Start()
	NewPeerPoolReader(pieceNo uint32, pieceLength uint32) io.ReadSeeker
}
type Impl struct {
	PeerFactory    peer.PeerFactory
	connectedHosts []domain.Host
	newHosts       []domain.Host
	connectedPeers []peer.Peer
}

var _ Service = &Impl{}

func (impl *Impl) AddHosts(newHosts ...domain.Host) {
	FilterNotChoking(impl.connectedPeers)
	impl.newHosts = append(impl.newHosts, newHosts...)
}

func (impl *Impl) Start() {
	fmt.Printf("starting peerpool\n")
	go impl.run()
}

func (impl *Impl) NewPeerPoolReader(pieceNo uint32, pieceLength uint32) io.ReadSeeker {
	//p := impl.connectedPeers[0]
	//return peer.NewPeerReader(p, pieceNo, pieceLength)
	return &poolReaderImpl{impl: impl,
		pieceNo:     pieceNo,
		pieceLength: pieceLength,
	}

}

type poolReaderImpl struct {
	impl        *Impl
	pieceNo     uint32
	pieceLength uint32
	curSeek     uint32
}

func (poolImpl *poolReaderImpl) Read(p []byte) (n int, err error) {
	filteredPeers := FilterPool(poolImpl.impl.connectedPeers, FilterNotChoking)
	if len(filteredPeers) == 0 {
		return 0, errors.New("no peers available")
	}
	r := peer.NewPeerReader(filteredPeers[0], poolImpl.pieceNo, poolImpl.pieceLength)
	r.Seek(int64(poolImpl.curSeek), io.SeekStart)
	return r.Read(p)
}

func (poolImpl *poolReaderImpl) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		poolImpl.curSeek += uint32(offset)
	case io.SeekStart:
		poolImpl.curSeek = uint32(offset)
	case io.SeekEnd:
		return 0, errors.New("seekend not implemented")
	}
	if poolImpl.curSeek > poolImpl.pieceLength {
		poolImpl.curSeek = poolImpl.pieceLength
	}
	return int64(poolImpl.curSeek), nil
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
