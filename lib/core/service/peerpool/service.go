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
	NewPeerPoolReader(pieceNo uint32, pieceLength uint32) io.ReadSeekCloser
}
type Impl struct {
	PeerFactory    peer.PeerFactory
	connectedHosts []domain.Host
	newHosts       []domain.Host
	connectedPeers []peer.Peer
}

var _ Service = &Impl{}

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

func (impl *Impl) NewPeerPoolReader(pieceNo uint32, pieceLength uint32) io.ReadSeekCloser {
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

func (poolImpl *poolReaderImpl) Read(p []byte) (int, error) {
Retry:
	filteredPeers := FilterPool(poolImpl.impl.connectedPeers, FilterNotChoking)
	if len(filteredPeers) == 0 {
		time.Sleep(1 * time.Second)
		fmt.Println("Waiting for peers")
		goto Retry
		//return 0, errors.New("no peers available")
	}
	targetPeer := filteredPeers[0]

	fmt.Printf("Choosing %s out of %d peers\n", targetPeer.GetPeerID(), len(filteredPeers))

	fmt.Println("Creating peer reader")
	r := peer.NewPeerReader(targetPeer, poolImpl.pieceNo, poolImpl.pieceLength)
	r.Seek(int64(poolImpl.curSeek), io.SeekStart)
	fmt.Println("Going to read")
	n, err := r.Read(p)
	fmt.Println("Done reading")
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		fmt.Println("error")
		return 0, err

	}
	poolImpl.curSeek += uint32(n)
	if err != nil {
		panic("err is not nil")
	}
	if n == 0 {
		panic("n is 0")
	}
	return n, nil
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
func (poolImpl *poolReaderImpl) Close() error {
	// TODO
	return nil
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
