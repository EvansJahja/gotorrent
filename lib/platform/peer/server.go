package peer

import (
	"fmt"
	"net"
	"strconv"

	"example.com/gotorrent/lib/core/adapter/peer"
	"example.com/gotorrent/lib/core/domain"
)

func (p PeerFactory) Serve(infoHash []byte) (chan peer.Peer, error) {
	newPeerChan := make(chan peer.Peer)
	startPort := 6881
	endPort := 6889
	for i := startPort; i < endPort; i++ {
		listenPort := ":" + strconv.Itoa(i)
		listen, err := net.Listen("tcp", listenPort)
		if err == nil {
			fmt.Printf("Listening on %s\n", listenPort)
			go p.acceptLoop(listen, infoHash, newPeerChan)
			return newPeerChan, nil
		}

	}
	panic("fail listening")
}

func (p PeerFactory) acceptLoop(l net.Listener, infoHash []byte, newPeerChan chan peer.Peer) {
	for {
		conn, err := l.Accept()
		fmt.Printf("Accept conn %s\n", conn.RemoteAddr().String())
		if err != nil {
			panic(err)
		}
		go func(conn net.Conn) {
			ipStr, portStr, err := net.SplitHostPort(conn.LocalAddr().String())
			if err != nil {
				panic(err)
			}
			port, err := strconv.Atoi(portStr)
			if err != nil {
				panic(err)
			}
			h := domain.Host{
				IP:   net.ParseIP(ipStr),
				Port: uint16(port),
			}
			impl := p.New(h).(*peerImpl)
			impl.conn = conn
			impl.handleConnection(conn)
			newPeerChan <- impl

		}(conn)
	}
}
