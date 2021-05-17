package peer

import (
	"fmt"
	"net"
	"strconv"

	"example.com/gotorrent/lib/core/adapter/peer"
	"example.com/gotorrent/lib/core/domain"
)

func (p PeerFactory) Serve(infoHash []byte) (newPeerChan chan peer.Peer, listenPort int, err error) {
	newPeerChan = make(chan peer.Peer)
	startPort := 6881
	endPort := 6889
	for i := startPort; i < endPort; i++ {
		listenPort = i
		listenPortStr := ":" + strconv.Itoa(i)
		var listen net.Listener
		listen, err = net.Listen("tcp", listenPortStr)
		if err == nil {
			fmt.Println(listen.Addr())
			fmt.Printf("Listening on %s\n", listenPortStr)
			go p.acceptLoop(listen, infoHash, newPeerChan)
			return
			//return newPeerChan, listenPort, nil
		}

	}
	return
}

func (p PeerFactory) acceptLoop(l net.Listener, infoHash []byte, newPeerChan chan peer.Peer) {
	for {
		conn, err := l.Accept()

		l_peer.Sugar().Debugf("accept conn from %s", conn.RemoteAddr().String())
		///fmt.Printf("Accept conn %s\n", conn.RemoteAddr().String())
		if err != nil {
			panic(err)
		}
		go func(conn net.Conn) {
			ipStr, portStr, err := net.SplitHostPort(conn.RemoteAddr().String())
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
			l_peer.Sugar().Debugf("accept conn from %s %s", impl.GetID(), conn.RemoteAddr().String())
			impl.conn = conn
			impl.handleConnection(conn)
			newPeerChan <- impl

		}(conn)
	}
}
