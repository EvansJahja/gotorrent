package peer

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"example.com/gotorrent/lib/core/adapter/peer"

	"example.com/gotorrent/lib/core/domain"
	"example.com/gotorrent/lib/extensions"
)

const (
	protoBitTorrent = "BitTorrent protocol"
)

type SendMsgFn func(msg []byte)

type peerImpl struct {
	Host         domain.Host
	Torrent      *domain.Torrent
	pieces       map[int]struct{}
	extHandler   extensions.ExtHandler
	conn         net.Conn
	notification []chan NotiMsg
}

type NotiMsg int

const (
	NotiInvalid NotiMsg = iota
	NotiPiecesUpdated
	NotiUnchocked
)

/*
type Peer interface {
	Connect() error
	GetHavePieces() map[int]struct{}
	GetNotification() <-chan NotiMsg

	RequestPiece(pieceId int)
}
*/

func New(h domain.Host) peer.Peer {
	p := peerImpl{
		Host:   h,
		pieces: make(map[int]struct{}),
	}
	return &p
}
func NewPeer(h domain.Host, t *domain.Torrent) peer.Peer {
	p := peerImpl{
		Host:    h,
		Torrent: t,
		pieces:  make(map[int]struct{}),
	}
	return &p
}

func (p *peerImpl) GetNotification() <-chan NotiMsg {
	n := make(chan NotiMsg)
	p.notification = append(p.notification, n)
	return n
}

func (p *peerImpl) Connect() error {

	hostname := net.JoinHostPort(p.Host.IP.String(), strconv.Itoa(int(p.Host.Port)))
	conn, err := net.DialTimeout("tcp", hostname, 3*time.Second)
	if err != nil {
		return err
	}
	p.conn = conn

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	p.Torrent.RLock()
	infoHash := p.Torrent.InfoHash
	p.Torrent.RUnlock()

	if err := doHandshake(conn, infoHash); err != nil {
		fmt.Print(err)
		return err
	}

	fmt.Printf("Connected to %s\n", hostname)

	go p.handleConn()

	return nil

}
func (p peerImpl) GetHavePieces() map[int]struct{} {
	return p.pieces
}

func (p peerImpl) RequestPiece(pieceId int) io.Reader {
	writeBuf := make([]byte, 12)

	binary.BigEndian.PutUint32(writeBuf[0:], uint32(pieceId))
	binary.BigEndian.PutUint32(writeBuf[4:], uint32(0))
	binary.BigEndian.PutUint32(writeBuf[8:], uint32(100))

	p.sendCmd(writeBuf, 6)

	//p.sendCmd()
	return bytes.NewBuffer([]byte{}) // TODO

}

func doHandshake(c net.Conn, infoHash string) error {
	handshakeReq := handshake{
		proto:        protoBitTorrent,
		featureFlags: 0x00_00_00_00_00_10_00_00,
		infoHash:     infoHash,
		peerID:       []byte("-GO0000-0257f4bc7fa1"),
	}
	c.Write(handshakeReq.getBytes())
	handshakeRespBuff := make([]byte, 68)

	n, err := c.Read(handshakeRespBuff)
	if err != nil {
		return err
	}
	handshakeRespBuff = handshakeRespBuff[:n]

	handshakeResp := newHandshake(handshakeRespBuff)
	if !handshakeResp.matches(handshakeReq) {
		return errors.New("handshake does not match")
	}

	return nil

}

func (h *peerImpl) handleConn() {

	h.extHandler = extensions.NewExtHandler(h.conn, h.sendCmd)

	h.extHandler.Init()

	go h.getMetadata()

	for {
		msgLenBuf := make([]byte, 4)
		_, err := h.conn.Read(msgLenBuf)
		if err != nil {
			break
		}
		h.conn.SetDeadline(time.Now().Add(5 * time.Minute))
		msgLen := binary.BigEndian.Uint32(msgLenBuf)

		if msgLen == 0 {
			// keep alive
			continue
		}

		msgBuf := make([]byte, msgLen)
		for n := 0; n < int(msgLen); {
			nextN, err := h.conn.Read(msgBuf[n:])
			if err != nil {
				goto exit
			}
			n += nextN
		}

		go h.handleMessage(msgBuf)
	}
exit:
}

func (h *peerImpl) handleMessage(msg []byte) {
	msgType := msg[0]
	fmt.Printf("Receive msg type %d\n", msgType)
	msgVal := msg[1:]

	switch msgType {
	case 1:
		print("Unchoke")
		h.notify(NotiUnchocked)
	case 5:
		print("Bitfield")
		h.sendCmd(nil, 2) // interested
		h.sendCmd(nil, 1) // unchoke
		h.handleBitField(msgVal)
		h.notify(NotiPiecesUpdated)
	case 20:
		h.handleExtendedMsg(msgVal)

	default:
		print(msgType)
	}
}

func (h *peerImpl) handleBitField(msg []byte) {
	for i, b := range msg {
		for j := 0; j < 8; j++ {
			key := i*8 + j
			val := (b >> j & 1) == 1
			if val {
				h.pieces[key] = struct{}{}
			}
		}
	}
}

func (h peerImpl) handleExtendedMsg(msg []byte) {
	extendedMsgId := msg[0]
	msgVal := msg[1:]

	h.extHandler.HandleCommand(extendedMsgId, msgVal)
}

func (h peerImpl) sendCmd(msg []byte, cmdId byte) {
	// Called for sending extended message
	msgLen := len(msg)

	// Send len + msg

	writeBuf := make([]byte, msgLen+4+1)
	binary.BigEndian.PutUint32(writeBuf[0:], uint32(msgLen+1))
	copy(writeBuf[4:], []byte{cmdId})
	copy(writeBuf[5:], msg)

	writer := bufio.NewWriter(h.conn)
	n, err := writer.Write(writeBuf)
	if err != nil {
		fmt.Println(err)
		fmt.Println(n)
	}
	writer.Flush()

	//h.c.Write(writeBuf)

}

func (h peerImpl) getMetadata() {
	if h.Torrent.Metadata != nil {
		return
	}
	metadata := <-h.extHandler.FetchMetadata()

	s := sha1.New()
	s.Write(metadata)
	b := s.Sum(nil)

	hash := hex.EncodeToString(b)
	if hash != h.Torrent.InfoHash {
		fmt.Printf("Invalid sum")
		return
	}

	h.Torrent.Lock()
	h.Torrent.Metadata = metadata
	fmt.Printf("%v", metadata)
	h.Torrent.Unlock()
	h.Torrent.HintUpdated()

}

func (p peerImpl) notify(msg NotiMsg) {
	go func() {
		for _, n := range p.notification {
			n <- msg
		}
	}()
}
