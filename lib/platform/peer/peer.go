package peer

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"example.com/gotorrent/lib/core/adapter/peer"

	"example.com/gotorrent/lib/core/domain"
	"example.com/gotorrent/lib/extensions"
)

const (
	protoBitTorrent = "BitTorrent protocol"
	maxRequestSize  = 1 << 14 // 2 ^ 14 or 16 kB
)

type SendMsgFn func(msg []byte)

type peerImpl struct {
	Host       domain.Host
	InfoHash   []byte
	pieces     map[int]struct{}
	extHandler extensions.ExtHandler
	conn       net.Conn

	weAreChocked      bool
	theyAreChocked    bool
	weAreInterested   bool
	theyAreInterested bool

	theirPeerID           []byte
	connected             bool
	peerHandshakeRespChan chan struct{}

	onChokedChangedFns []func(bool)
	onPiecesChangedFns []func()
	onPieceArriveFns   []func(index uint32, begin uint32, piece []byte)

	internalOnPieceArriveChans sync.Map
	//notificationMut    sync.RWMutex
}

type keyType struct {
	pieceId uint32
	begin   uint32
	length  uint32
}

func New(h domain.Host, infoHash []byte) peer.Peer {
	p := peerImpl{
		InfoHash:              infoHash,
		Host:                  h,
		pieces:                make(map[int]struct{}),
		peerHandshakeRespChan: make(chan struct{}),

		weAreChocked:      true,
		theyAreChocked:    true,
		weAreInterested:   false,
		theyAreInterested: false,
	}
	return &p
}

var _ peer.Peer = &peerImpl{}

func (impl *peerImpl) GetState() peer.State {
	return peer.State{
		WeAreChocked:      impl.weAreChocked,
		TheyAreChocked:    impl.theyAreChocked,
		WeAreInterested:   impl.weAreInterested,
		TheyAreInterested: impl.theyAreInterested,
		Connected:         impl.connected,
	}

}

func (impl *peerImpl) GetPeerID() []byte {
	<-impl.peerHandshakeRespChan
	return impl.theirPeerID
}

func (impl *peerImpl) Connect() error {

	hostname := net.JoinHostPort(impl.Host.IP.String(), strconv.Itoa(int(impl.Host.Port)))
	conn, err := net.DialTimeout("tcp", hostname, 3*time.Second)
	if err != nil {
		//fmt.Printf("Fail connecting to %s, err: %s\n", hostname, err.Error())
		return err
	}
	impl.conn = conn

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	if err := impl.doHandshake(); err != nil {
		//fmt.Printf("Fail connecting to %s, err: %s\n", hostname, err.Error())
		return err
	}

	fmt.Printf("Connected to %s\n", hostname)
	impl.connected = true

	go impl.handleConn()

	return nil

}
func (impl *peerImpl) GetHavePieces() map[int]struct{} {
	return impl.pieces
}

func (impl *peerImpl) RequestPiece(pieceId uint32, begin uint32, length uint32) {
	writeBuf := make([]byte, 12)
	if length > maxRequestSize {
		length = maxRequestSize
	}

	binary.BigEndian.PutUint32(writeBuf[0:], uint32(pieceId))
	binary.BigEndian.PutUint32(writeBuf[4:], uint32(begin))
	binary.BigEndian.PutUint32(writeBuf[8:], uint32(length))

	impl.sendCmd(writeBuf, 6)

}

func (impl *peerImpl) RequestPieceWithChan(pieceId uint32, begin uint32, length uint32) <-chan []byte {
	resultCh := make(chan []byte, 1)
	key := keyType{
		pieceId: pieceId,
		begin:   begin,
		length:  length,
	}

	impl.internalOnPieceArriveChans.Store(key, resultCh)
	impl.RequestPiece(pieceId, begin, length)

	return resultCh
}

func (impl *peerImpl) doHandshake() error {
	handshakeReq := handshake{
		proto:        protoBitTorrent,
		featureFlags: 0x00_00_00_00_00_10_00_00,
		infoHash:     impl.InfoHash,
		peerID:       []byte("-GO0000-0257f4bc7fa1"),
	}
	impl.conn.Write(handshakeReq.getBytes())
	handshakeRespBuff := make([]byte, 68)

	n, err := impl.conn.Read(handshakeRespBuff)
	if err != nil {
		return err
	}
	handshakeRespBuff = handshakeRespBuff[:n]

	handshakeResp := newHandshake(handshakeRespBuff)
	if !handshakeResp.matches(handshakeReq) {
		return errors.New("handshake does not match")
	}

	go func() {
		impl.theirPeerID = handshakeResp.peerID
		close(impl.peerHandshakeRespChan)
	}()

	return nil

}

func (impl *peerImpl) handleConn() {

	impl.extHandler = extensions.NewExtHandler(impl.conn, impl.sendCmd)

	impl.extHandler.Init()

	///impl.extHandler.Startup()
	// TODO: We don't need this now
	//go impl.getMetadata()

	for {
		msgLenBuf := make([]byte, 4)
		_, err := impl.conn.Read(msgLenBuf)
		if err != nil {
			break
		}
		impl.conn.SetDeadline(time.Now().Add(5 * time.Minute))
		msgLen := binary.BigEndian.Uint32(msgLenBuf)

		if msgLen == 0 {
			// keep alive
			continue
		}

		msgBuf := make([]byte, msgLen)
		for n := 0; n < int(msgLen); {
			nextN, err := impl.conn.Read(msgBuf[n:])
			if err != nil {
				goto exit
			}
			n += nextN
		}

		go impl.handleMessage(msgBuf)
	}
exit:
	fmt.Printf("Dead: %s\n", impl.theirPeerID)
	impl.connected = false
}

func (impl *peerImpl) handleMessage(msg []byte) {
	msgType := msg[0]
	fmt.Printf("Receive msg type %d\n", msgType)
	msgVal := msg[1:]

	switch msgType {
	case 0: // Choke
		impl.handleWeAreChoked(true)
	case 1: // Unchoke
		impl.handleWeAreChoked(false)
	case 5:
		print("Bitfield")
		impl.handleBitField(msgVal)
		//impl.notify(NotiPiecesUpdated)
	case 7:
		impl.handlePiece(msgVal)
	case 20:
		impl.handleExtendedMsg(msgVal)

	default:
		print(msgType)
	}
}

func (impl *peerImpl) handleWeAreChoked(weAreChoked bool) {
	impl.weAreChocked = weAreChoked
	var wg sync.WaitGroup
	for _, onChokedChanged := range impl.onChokedChangedFns {
		wg.Add(1)
		go func(f func(b bool)) {
			f(weAreChoked)
			wg.Done()
		}(onChokedChanged)
	}
	wg.Wait()
}

func (impl *peerImpl) handlePiece(msg []byte) {
	var index uint32
	var begin uint32
	var piece []byte

	index = binary.BigEndian.Uint32(msg[0:4])
	begin = binary.BigEndian.Uint32(msg[4:8])
	piece = msg[8:]

	key := keyType{
		pieceId: index,
		begin:   begin,
		length:  uint32(len(piece)),
	}
	if chInterface, ok := impl.internalOnPieceArriveChans.LoadAndDelete(key); ok {
		ch := chInterface.(chan []byte)
		ch <- piece
	}

	var wg sync.WaitGroup
	for _, onPieceArriveFn := range impl.onPieceArriveFns {
		wg.Add(1)
		go func(f func(i uint32, b uint32, p []byte)) {
			f(index, begin, piece)
			wg.Done()
		}(onPieceArriveFn)
	}
	wg.Wait()
}

func (impl *peerImpl) handleBitField(msg []byte) {
	for i, b := range msg {
		for j := 0; j < 8; j++ {
			key := i*8 + j
			val := (b >> j & 1) == 1
			if val {
				impl.pieces[key] = struct{}{}
			}
		}
	}

	var wg sync.WaitGroup
	for _, piecesChangedNotif := range impl.onPiecesChangedFns {
		wg.Add(1)
		go func(f func()) {
			f()
			wg.Done()
		}(piecesChangedNotif)
	}
	wg.Wait()

}

func (impl peerImpl) handleExtendedMsg(msg []byte) {
	extendedMsgId := msg[0]
	msgVal := msg[1:]

	impl.extHandler.HandleCommand(extendedMsgId, msgVal)
}

func (impl peerImpl) sendCmd(msg []byte, cmdId byte) {
	// Called for sending extended message
	msgLen := len(msg)

	// Send len + msg

	writeBuf := make([]byte, msgLen+4+1)
	binary.BigEndian.PutUint32(writeBuf[0:], uint32(msgLen+1))
	copy(writeBuf[4:], []byte{cmdId})
	copy(writeBuf[5:], msg)

	writer := impl.conn
	n, err := writer.Write(writeBuf)
	if err != nil {
		fmt.Println(err)
		fmt.Println(n)
	}

	//h.c.Write(writeBuf)

}

func (impl peerImpl) GetMetadata() (domain.Metadata, error) {
	metadataBytes := <-impl.extHandler.FetchMetadata()

	metadata := domain.Metadata(metadataBytes)

	if !bytes.Equal(metadata.InfoHash(), impl.InfoHash) {
		return nil, errors.New("Invalid sum")
	}

	return domain.Metadata(metadata), nil

}

/*
func (impl peerImpl) notify(msg NotiMsg) {
	go func() {
		for _, n := range impl.notification {
			n <- msg
		}
	}()
}
*/

func (impl *peerImpl) Choke() {}
func (impl *peerImpl) Unchoke() {
	impl.sendCmd(nil, 1) // unchoke
}
func (impl *peerImpl) Interested() {
	impl.sendCmd(nil, 2) // interested
}
func (impl *peerImpl) Uninterested() {}
