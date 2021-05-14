package udptracker

/*
 * BEP 15 UDPTracker
 */
import (
	"encoding/binary"
	"fmt"
	"net"
	"net/url"
	"time"

	"example.com/gotorrent/lib/core/adapter/clock"
	"example.com/gotorrent/lib/core/adapter/peerlist"
	"example.com/gotorrent/lib/core/domain"
)

type UdpPeerList struct {
	InfoHash []byte
	Trackers []*url.URL
	Clock    clock.Clock
}

var _ peerlist.PeerRepo = UdpPeerList{}

func (peerList UdpPeerList) Start() {}
func (peerList UdpPeerList) Stop()  {}

func (peerList UdpPeerList) GetPeers() []domain.Host {
	trackerURLs := peerList.Trackers
	var hosts []domain.Host
	for _, t := range trackerURLs {

		resp, err := peerList.announce(t)
		if err == nil {
			hosts = append(hosts, resp.Hosts...)
		}

	}
	return hosts
}

type connectRequest struct {
	protocolId    uint64
	action        uint32
	transactionId uint32
}

func newConnectRequest() connectRequest {
	u := connectRequest{}
	u.protocolId = 0x41727101980
	u.action = 0x0
	return u
}

func (u connectRequest) getBytes() []byte {
	b := make([]byte, 16)

	binary.BigEndian.PutUint64(b[0:], u.protocolId)
	binary.BigEndian.PutUint32(b[8:], u.action)
	binary.BigEndian.PutUint32(b[12:], u.transactionId)

	return b

}

type connectResponse struct {
	action        uint32
	transactionId uint32
	connID        uint64
}

func (u connectResponse) matchesWithReq(v connectRequest) bool {
	return u.action == v.action && u.transactionId == v.transactionId
}
func newConnectResponse(b []byte) connectResponse {
	var u connectResponse
	u.action = binary.BigEndian.Uint32(b[0:])
	u.transactionId = binary.BigEndian.Uint32(b[4:])
	u.connID = binary.BigEndian.Uint64(b[8:])
	return u
}

type announceRequest struct {
	connId        uint64
	action        uint32
	transactionId uint32
	infoHash      []byte
	peerId        string
	downloaded    uint64
	left          uint64
	uploaded      uint64
	event         uint32
	ip            uint32
	key           uint32
	num_want      uint32
	port          uint16
}

func (u announceRequest) getBytes() []byte {
	b := make([]byte, 98)
	binary.BigEndian.PutUint64(b[0:], u.connId)
	binary.BigEndian.PutUint32(b[8:], u.action)
	binary.BigEndian.PutUint32(b[12:], u.transactionId)

	copy(b[16:16+20], []byte(u.infoHash))
	copy(b[36:36+20], []byte(u.peerId))

	binary.BigEndian.PutUint64(b[56:], u.downloaded)
	binary.BigEndian.PutUint64(b[64:], u.left)
	binary.BigEndian.PutUint64(b[72:], u.uploaded)
	binary.BigEndian.PutUint32(b[80:], u.event)
	binary.BigEndian.PutUint32(b[84:], u.ip)
	binary.BigEndian.PutUint32(b[88:], u.key)
	binary.BigEndian.PutUint32(b[92:], u.num_want)
	binary.BigEndian.PutUint16(b[96:], u.port)
	return b
}

func newAnnounceRequest() announceRequest {
	var u announceRequest
	u.action = 0x1
	u.ip = 0
	num_want := -1
	u.num_want = uint32(num_want)
	u.event = 0
	return u
}

/*
type Host struct {
	IP   net.IP
	Port uint16
}
*/

type AnnounceResponse struct {
	Action   uint32
	TxnID    uint32
	Interval uint32
	Leechers uint32
	Seeders  uint32
	Hosts    []domain.Host
}

func newAnnounceResponse(b []byte) AnnounceResponse {
	var u AnnounceResponse

	u.Action = binary.BigEndian.Uint32(b[0:])
	u.TxnID = binary.BigEndian.Uint32(b[4:])
	u.Interval = binary.BigEndian.Uint32(b[8:])
	u.Leechers = binary.BigEndian.Uint32(b[12:])

	n := (len(b) - 20) / 6

	if (len(b)-20)%6 != 0 {
		panic("wrong")
	}

	for i := 0; i < n; i++ {
		ip := net.IP(b[20+6*i : 20+6*i+4])
		port := binary.BigEndian.Uint16(b[24+6*i : 24+6*i+4])

		newHost := domain.Host{IP: ip, Port: port}
		// Note: This is in place due to announce having Port 0
		if newHost.Port != 0 {
			u.Hosts = append(u.Hosts, newHost)
		}

	}

	return u
}

// Not stable
func (peerList UdpPeerList) Announce(u *url.URL) (AnnounceResponse, error) {
	return peerList.announce(u)
}

func (peerList UdpPeerList) announce(u *url.URL) (AnnounceResponse, error) {

	transactionId := uint32(peerList.Clock.Now().UnixNano() ^ 0xdeadbeef)

	connReq := newConnectRequest()
	connReq.transactionId = transactionId

	c, err := net.Dial("udp", u.Host)
	if err != nil {
		return AnnounceResponse{}, err
	}
	defer c.Close()
	if err != nil {
		fmt.Print(err)
		return AnnounceResponse{}, err
	}
	c.SetDeadline(peerList.Clock.Now().Add(3 * time.Second))
	bytesToWrite := connReq.getBytes()

	c.Write(bytesToWrite)

	buffRead := make([]byte, 32)

	count, err := c.Read(buffRead)
	if err != nil {
		return AnnounceResponse{}, err
	}
	buffRead = buffRead[:count]

	connResp := newConnectResponse(buffRead)
	if !connResp.matchesWithReq(connReq) {
		return AnnounceResponse{}, err
	}

	//connId := connResp.connID

	announceReq := newAnnounceRequest()

	announceReq.connId = connResp.connID
	announceReq.transactionId = transactionId
	announceReq.infoHash = peerList.InfoHash
	announceReq.peerId = "-GO0000-0257f4bc7fa1"

	c.Write(announceReq.getBytes())

	buffRead = make([]byte, 4096)

	l, err := c.Read(buffRead)
	if err != nil {
		return AnnounceResponse{}, err
	}
	buffRead = buffRead[:l]

	announceResp := newAnnounceResponse(buffRead)

	/* example resp
	* {Action:1 TxnID:2340680293 Interval:1697 Leechers:8 Seeders:0 Hosts: [...]
	 */
	//fmt.Printf("%+v", announceResp)
	return announceResp, nil

}
