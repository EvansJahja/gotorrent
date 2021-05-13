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
	"example.com/gotorrent/lib/extensions"

	"example.com/gotorrent/lib/core/domain"
)

const (
	protoBitTorrent = "BitTorrent protocol"
	maxRequestSize  = 1 << 14 // 2 ^ 14 or 16 kB
)

type SendMsgFn func(msg []byte)

type MsgType byte

const (
	MsgChoke         MsgType = 0
	MsgUnchoke       MsgType = 1
	MsgInterested    MsgType = 2
	MsgNotInterested MsgType = 3
	MsgHave          MsgType = 4
	MsgBitfield      MsgType = 5
	MsgRequest       MsgType = 6
	MsgPiece         MsgType = 7
	MsgCancel        MsgType = 8
	MsgExtended      MsgType = 20
)

type peerImpl struct {
	Host        domain.Host
	InfoHash    []byte
	theirPieces domain.PieceList
	ourPiecesFn func() domain.PieceList
	extHandler  extensions.ExtHandler
	conn        net.Conn

	weAreChocked      bool
	theyAreChocked    bool
	weAreInterested   bool
	theyAreInterested bool

	theirPeerID           []byte
	connected             bool
	peerHandshakeRespChan chan struct{}

	onChokedChangedFns []func(bool)
	onPiecesChangedFns []func()

	internalOnPieceArriveChans sync.Map
	pieceRequestChan           chan peer.PieceRequest
	//notificationMut    sync.RWMutex
}

var _ peer.Peer = &peerImpl{}

type keyType struct {
	pieceId uint32
	begin   uint32
	length  uint32
}

type PeerFactory struct {
	InfoHash       []byte
	OurPieceListFn func() domain.PieceList
}

func (p PeerFactory) New(host domain.Host) peer.Peer {
	newPeer := new(host, p.InfoHash).(*peerImpl)
	newPeer.ourPiecesFn = p.OurPieceListFn
	return newPeer
}

func new(h domain.Host, infoHash []byte) peer.Peer {
	p := peerImpl{
		InfoHash:              infoHash,
		Host:                  h,
		peerHandshakeRespChan: make(chan struct{}),

		weAreChocked:      true,
		theyAreChocked:    true,
		weAreInterested:   false,
		theyAreInterested: false,
		pieceRequestChan:  make(chan peer.PieceRequest),
	}
	return &p
}

var _ peer.Peer = &peerImpl{}

func (impl *peerImpl) Hostname() string {
	return impl.Host.IP.String()
}

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
	//<-impl.peerHandshakeRespChan
	return impl.theirPeerID
}

func (impl *peerImpl) Connect() error {
	hostname := net.JoinHostPort(impl.Host.IP.String(), strconv.Itoa(int(impl.Host.Port)))
	conn, err := net.DialTimeout("tcp", hostname, 3*time.Second)
	if err != nil {
		//fmt.Printf("Fail connecting to %s, err: %s\n", hostname, err.Error())
		return err
	}
	return impl.handleConnection(conn)
}

func (impl *peerImpl) handleConnection(conn net.Conn) error {
	hostname := net.JoinHostPort(impl.Host.IP.String(), strconv.Itoa(int(impl.Host.Port)))

	impl.conn = conn

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	if err := impl.doHandshake(); err != nil {
		//fmt.Printf("Fail connecting to %s, err: %s\n", hostname, err.Error())
		return err
	}

	fmt.Printf("Connected to %s\n", hostname)
	impl.connected = true

	impl.conn.SetDeadline(time.Now().Add(5 * time.Minute))
	go impl.handleConn()

	return nil

}

func (impl *peerImpl) TheirPieces() domain.PieceList {
	return impl.theirPieces
}

func (impl *peerImpl) OurPieces() domain.PieceList {
	return impl.ourPiecesFn()
}

func (impl *peerImpl) SetOurPiece(pieceNo uint32) {
	panic("implement me")

}

func (impl *peerImpl) RequestPiece(pieceId uint32, begin uint32, length uint32) <-chan []byte {
	if length > maxRequestSize {
		length = maxRequestSize
	}

	resultCh := make(chan []byte, 1)
	key := keyType{
		pieceId: pieceId,
		begin:   begin,
		length:  length,
	}

	impl.internalOnPieceArriveChans.Store(key, resultCh)

	writeBuf := make([]byte, 12)

	binary.BigEndian.PutUint32(writeBuf[0:], uint32(pieceId))
	binary.BigEndian.PutUint32(writeBuf[4:], uint32(begin))
	binary.BigEndian.PutUint32(writeBuf[8:], uint32(length))

	impl.sendCmd(writeBuf, MsgRequest)

	return resultCh
}

func (impl *peerImpl) PieceRequests() <-chan peer.PieceRequest {
	return impl.pieceRequestChan
}

func (impl *peerImpl) doHandshake() error {
	handshakeReq := handshake{
		proto:        protoBitTorrent,
		featureFlags: 0x00_00_00_00_00_10_00_00,
		infoHash:     impl.InfoHash,
		peerID:       []byte("-GO0000-0257f4bc7fa1"),
	}

	fmt.Println("Send handshake")
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

	impl.theirPeerID = handshakeResp.peerID
	//close(impl.peerHandshakeRespChan)
	/*
		go func() {
		}()
	*/

	return nil

}

func (impl *peerImpl) handleConn() {
	sendCmdUntyped := extensions.SendMsgFn(func(msg []byte, msgType byte) {
		impl.sendCmd(msg, MsgType(msgType))
	})

	impl.extHandler = extensions.NewExtHandler(impl.conn, sendCmdUntyped)

	impl.extHandler.Init()
	impl.sendBitfields()

	///impl.extHandler.Startup()
	// TODO: We don't need this now
	//go impl.getMetadata()

	for {
		msgLenBuf := make([]byte, 4)
		_, err := impl.read(msgLenBuf)
		if err != nil {
			impl.disconnected(err)
			return
		}
		impl.conn.SetDeadline(time.Now().Add(5 * time.Minute))
		msgLen := binary.BigEndian.Uint32(msgLenBuf)

		if msgLen == 0 {
			// keep alive
			continue
		}

		msgBuf := make([]byte, msgLen)
		for n := 0; n < int(msgLen); {
			nextN, err := impl.read(msgBuf[n:])
			if err != nil {
				impl.disconnected(err)
				return
			}
			n += nextN
		}

		go impl.handleMessage(msgBuf)
	}
}

func (impl *peerImpl) handleMessage(msg []byte) {
	msgType := MsgType(msg[0])
	msgVal := msg[1:]

	switch msgType {
	case MsgChoke:
		impl.handleWeAreChoked(true)
	case MsgUnchoke:
		impl.handleWeAreChoked(false)
	case MsgInterested:
		impl.handleTheyAreInterested(true)
	case MsgBitfield:
		print("Bitfield")
		impl.handleBitField(msgVal)
	case MsgRequest:
		impl.handleRequest(msgVal)
	case MsgPiece:
		impl.handlePiece(msgVal)
	case MsgExtended:
		impl.handleExtendedMsg(msgVal)

	default:
		print(msgType)
	}
}

func (impl *peerImpl) handleWeAreChoked(weAreChoked bool) {
	if weAreChoked {
		fmt.Printf("%s choked\n", impl.Host.IP.String())
	} else {
		fmt.Printf("%s unchoked\n", impl.Host.IP.String())
	}
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
func (impl *peerImpl) handleTheyAreInterested(interested bool) {
	impl.Unchoke()

}

func (impl *peerImpl) handleRequest(msg []byte) {

	pieceId := binary.BigEndian.Uint32(msg[0:])
	begin := binary.BigEndian.Uint32(msg[4:])
	length := binary.BigEndian.Uint32(msg[8:])
	if !impl.ourPiecesFn().ContainPiece(pieceId) {
		return
	}

	fmt.Printf("req #%d %d %d\n", pieceId, begin, length)
	respCh := make(chan []byte)
	req := peer.PieceRequest{
		PieceNo:  pieceId,
		Begin:    begin,
		Length:   length,
		Response: respCh,
	}
	go func() {
		impl.pieceRequestChan <- req
		resp := <-respCh
		impl.sendPiece(pieceId, begin, resp)
	}()

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
	if chInterface, ok := impl.internalOnPieceArriveChans.Load(key); ok {
		ch := chInterface.(chan []byte)
		ch <- piece
	} else {
		panic("no handler")
	}
}

func (impl *peerImpl) handleBitField(msg []byte) {
	/*
		for i, b := range msg {
			for j := 0; j < 8; j++ {
				key := i*8 + j
				val := (b >> j & 1) == 1
				if val {
					impl.theirPieces[key] = struct{}{}
				}
			}
		}
	*/
	impl.theirPieces = domain.PieceList(msg)

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

func (impl *peerImpl) handleExtendedMsg(msg []byte) {
	extendedMsgId := msg[0]
	msgVal := msg[1:]

	impl.extHandler.HandleCommand(extendedMsgId, msgVal)
}

func (impl *peerImpl) sendCmd(msg []byte, msgType MsgType) {
	// Called for sending extended message
	msgLen := len(msg)

	// Send len + msg

	writeBuf := make([]byte, msgLen+4+1)
	binary.BigEndian.PutUint32(writeBuf[0:], uint32(msgLen+1))
	copy(writeBuf[4:], []byte{byte(msgType)})
	copy(writeBuf[5:], msg)

	_, err := impl.write(writeBuf)
	if err != nil {
		impl.disconnected(err)
	}

	//h.c.Write(writeBuf)
}

func (impl *peerImpl) sendBitfields() {
	impl.sendCmd([]byte(impl.ourPiecesFn()), MsgBitfield)
}

func (impl *peerImpl) sendPiece(pieceNo uint32, begin uint32, piece []byte) {
	buf := make([]byte, len(piece)+4+4)
	binary.BigEndian.PutUint32(buf[0:], pieceNo)
	binary.BigEndian.PutUint32(buf[4:], begin)
	copy(buf[8:], piece)
	impl.sendCmd(buf, MsgPiece)

}

func (impl *peerImpl) GetMetadata() (domain.Metadata, error) {
	metadataBytes := <-impl.extHandler.FetchMetadata()

	metadata := domain.Metadata(metadataBytes)

	if !bytes.Equal(metadata.InfoHash(), impl.InfoHash) {
		return nil, errors.New("invalid sum")
	}

	return domain.Metadata(metadata), nil

}

func (impl *peerImpl) Choke() {
	impl.sendCmd(nil, MsgChoke)
}
func (impl *peerImpl) Unchoke() {
	impl.sendCmd(nil, MsgUnchoke)
}
func (impl *peerImpl) Interested() {
	impl.sendCmd(nil, MsgInterested)
}
func (impl *peerImpl) Uninterested() {
	impl.sendCmd(nil, MsgNotInterested)
}

func (impl *peerImpl) disconnected(reason error) {
	if !impl.connected {
		return
	}
	fmt.Printf("Dead: %s\n", impl.theirPeerID)
	impl.connected = false
	impl.conn.Close()
}

func (impl *peerImpl) read(b []byte) (int, error) {
	if !impl.connected {
		return 0, errors.New("disconnected")
	}
	n, err := impl.conn.Read(b)
	if err != nil {
		impl.disconnected(err)
	}
	return n, err
}
func (impl *peerImpl) write(b []byte) (int, error) {
	if !impl.connected {
		return 0, errors.New("disconnected")
	}
	n, err := impl.conn.Write(b)
	if err != nil {
		impl.disconnected(err)

	}
	return n, err
}
