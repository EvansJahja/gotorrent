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
)

type Service interface {
	AddHosts(newHosts ...domain.Host)
	Start()
	//	NewPeerPoolReader(pieceNo uint32, pieceLength uint32) io.ReadSeekCloser
	NewPeerPoolReader(pieceNo uint32, pieceLength int, pieceCount int, torrentLength int) io.ReadSeekCloser
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

func (impl *Impl) NewPeerPoolReader(pieceNo uint32, pieceLength int, pieceCount int, torrentLength int) io.ReadSeekCloser {
	return &poolReaderImpl{impl: impl,
		pieceNo:       pieceNo,
		pieceLength:   pieceLength,
		pieceCount:    pieceCount,
		torrentLength: torrentLength,
	}

}

type poolReaderImpl struct {
	impl          *Impl
	pieceNo       uint32
	pieceLength   int
	curSeek       int
	pieceCount    int
	torrentLength int
}

func (poolImpl *poolReaderImpl) Read(p []byte) (int, error) {
Retry:
	filteredPeers := FilterPool(poolImpl.impl.connectedPeers, FilterConnected, FilterNotChoking)
	if len(filteredPeers) == 0 {
		time.Sleep(1 * time.Second)
		goto Retry
		//return 0, errors.New("no peers available")
	}
	//peerIdx := 0
	peerIdx := rand.Int() % len(filteredPeers)
	targetPeer := filteredPeers[peerIdx]

	fmt.Printf("Choosing %s out of %d peers\n", targetPeer.GetPeerID(), len(filteredPeers))

	fmt.Println("Creating peer reader")

	// Limit reading to piece length
	pieceLength := poolImpl.pieceLength
	if poolImpl.pieceNo == uint32(poolImpl.pieceCount-1) {
		// Last
		lastRemainder := poolImpl.torrentLength % poolImpl.pieceLength
		if lastRemainder != 0 {
			pieceLength = lastRemainder
		}
	}
	remainingToRead := pieceLength - poolImpl.curSeek
	bufClamp := len(p)
	if bufClamp > remainingToRead {
		bufClamp = remainingToRead
		if bufClamp < 0 {
			return 0, errors.New("invalid read pos")
		}
	}
	p = p[:bufClamp] // Clamp to remainingToRead

	r := peer.NewPeerReader(targetPeer, poolImpl.pieceNo)
	r.Seek(int64(poolImpl.curSeek), io.SeekStart)
	fmt.Println("Going to read")
	n, err := r.Read(p)
	fmt.Println("Done reading")
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		fmt.Println("error")
		return 0, err

	}
	poolImpl.curSeek += n
	if err != nil {
		panic("err is not nil")
	}
	if n == 0 {
		panic("n is 0")
	}
	return n, nil
}

func (poolImpl *poolReaderImpl) Seek(offset int64, whence int) (int64, error) {
	// Dup code

	pieceLength := poolImpl.pieceLength
	if poolImpl.pieceNo == uint32(poolImpl.pieceCount-1) {
		// Last
		lastRemainder := poolImpl.torrentLength % poolImpl.pieceLength
		if lastRemainder != 0 {
			pieceLength = lastRemainder
		}
	}
	// End of Dup code

	switch whence {
	case io.SeekCurrent:
		poolImpl.curSeek += int(offset)
	case io.SeekStart:
		poolImpl.curSeek = int(offset)
	case io.SeekEnd:
		return 0, errors.New("seekend not implemented")
	}
	if poolImpl.curSeek > pieceLength {
		poolImpl.curSeek = pieceLength
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
