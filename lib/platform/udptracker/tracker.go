package udptracker

import (
	"context"
	"encoding/binary"
	"net"
	"net/url"
	"time"

	"example.com/gotorrent/lib/core/domain"
	"example.com/gotorrent/lib/logger"
)

// Handle individual tracker, as well as re-requesting

var l_tracker = l_udptracker.Named("trackerImpl")

type tracker struct {
	trackerUrl  *url.URL
	newHostChan chan domain.Host
	infoHash    []byte
	connID      connectionID
	n           int // for retrying delay. See BEP 0015
	isConnected bool
	trackerInfo TrackerInfo
}

type connectionID struct {
	id      uint64
	gotTime time.Time
}

func (c connectionID) expired() bool {
	if (c.gotTime == time.Time{}) {
		return true
	}

	expiryTime := c.gotTime.Add(60 * time.Second)

	return time.Now().After(expiryTime)

}

func newConnectionID(id uint64) connectionID {
	return connectionID{
		id:      id,
		gotTime: time.Now(),
	}
}

func (t *tracker) updateInfo(newInfo TrackerInfo) {
	t.trackerInfo = newInfo
}

func (t *tracker) run() {
	ctx := logger.NewContextid(context.Background())
	l := logger.Ctx(l_tracker, ctx)
	for {
		// connection ID lives for 1 minute

		if t.connID.expired() {
			l.Sugar().Debugw("conn ID expired", "connID", t.connID.id)
			t.isConnected = false
		} else {
			if t.isConnected {
				goto Connected
			}
		}

		{
			connResp, err := t.connect(ctx, t.trackerUrl, getTimeout(t.n))
			if err != nil {
				//if errors.Is(err, os.ErrDeadlineExceeded) {
				if t.n+1 < 8 {
					t.n++
				}
				l.Sugar().Warnw("fail to connect, retrying", "n", t.n)
				continue
			}
			t.n = 0
			t.connID = newConnectionID(connResp.connID)
			t.isConnected = true
			l.Sugar().Infow("connected to tracker", "connID", t.connID.id, "url", t.trackerUrl)
		}

	Connected:

		announceResp, err := t.announce(ctx, t.trackerUrl, getTimeout(t.n))
		if err != nil {
			if t.n+1 < 8 {
				t.n++
			}
			continue
		}

		announceInterval := time.Duration(announceResp.Interval) * time.Second

		for _, h := range announceResp.Hosts {
			l.Sugar().Debugw("got host from announce", "host", h)
			t.newHostChan <- h
		}

		<-time.After(announceInterval)

	}

}

func getTimeout(retryN int) time.Duration {
	// For other than connection, timeout should follow spec
	n := retryN
	if n >= 8 {
		n = 8
	}
	return time.Duration(15*1<<n) * time.Second

}

type connectRequest struct {
	protocolId    uint64
	action        uint32
	transactionID uint32
}

func newConnectRequest() connectRequest {
	u := connectRequest{}
	u.protocolId = 0x41727101980
	u.action = 0x0
	return u
}

func newConnectRequestFromBytes(b []byte) connectRequest {
	r := newConnectRequest()

	r.protocolId = binary.BigEndian.Uint64(b[0:])
	r.action = binary.BigEndian.Uint32(b[8:])
	r.transactionID = binary.BigEndian.Uint32(b[12:])

	return r

}

func (u connectRequest) getBytes() []byte {
	b := make([]byte, 16)

	binary.BigEndian.PutUint64(b[0:], u.protocolId)
	binary.BigEndian.PutUint32(b[8:], u.action)
	binary.BigEndian.PutUint32(b[12:], u.transactionID)

	return b

}

type connectResponse struct {
	action        uint32
	transactionId uint32
	connID        uint64
}

func (u connectResponse) matchesWithReq(v connectRequest) bool {
	return u.action == v.action && u.transactionId == v.transactionID
}
func (u connectResponse) getBytes() []byte {
	b := make([]byte, 16)
	binary.BigEndian.PutUint32(b[0:], u.action)
	binary.BigEndian.PutUint32(b[4:], u.transactionId)
	binary.BigEndian.PutUint64(b[8:], u.connID)
	return b
}

func newConnectResponse(b []byte) connectResponse {
	var u connectResponse
	u.action = binary.BigEndian.Uint32(b[0:])
	u.transactionId = binary.BigEndian.Uint32(b[4:])
	u.connID = binary.BigEndian.Uint64(b[8:])
	return u
}

type announceRequest struct {
	ConnID        uint64
	Action        uint32
	TransactionID uint32
	InfoHash      []byte
	PeerID        string
	Downloaded    uint64
	Left          uint64
	Uploaded      uint64
	Event         uint32
	IP            uint32
	Key           uint32
	NumWant       int32
	Port          uint16
}

func (u announceRequest) getBytes() []byte {
	b := make([]byte, 98)
	binary.BigEndian.PutUint64(b[0:], u.ConnID)
	binary.BigEndian.PutUint32(b[8:], u.Action)
	binary.BigEndian.PutUint32(b[12:], u.TransactionID)

	copy(b[16:16+20], []byte(u.InfoHash))
	copy(b[36:36+20], []byte(u.PeerID))

	binary.BigEndian.PutUint64(b[56:], u.Downloaded)
	binary.BigEndian.PutUint64(b[64:], u.Left)
	binary.BigEndian.PutUint64(b[72:], u.Uploaded)
	binary.BigEndian.PutUint32(b[80:], u.Event)
	binary.BigEndian.PutUint32(b[84:], u.IP)
	binary.BigEndian.PutUint32(b[88:], u.Key)
	binary.BigEndian.PutUint32(b[92:], uint32(u.NumWant))
	binary.BigEndian.PutUint16(b[96:], u.Port)
	return b
}

func newAnnounceRequest() announceRequest {
	var u announceRequest
	u.Action = 0x1
	u.IP = 0
	num_want := -1
	u.NumWant = int32(num_want)
	u.Event = 0
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

func (impl *tracker) connect(ctx context.Context, u *url.URL, readTimeout time.Duration) (connectResponse, error) {
	l := logger.Ctx(l_tracker, ctx).Named("connect").Sugar()

	connReq := newConnectRequest()
	connReq.transactionID = newTransactionID()

	l.Debugw("dialing", "host", u.Host)
	c, err := net.Dial("udp", u.Host)
	if err != nil {
		l.Errorw("error dial", "host", u.Host, "err", err.Error())
		return connectResponse{}, err
	}
	defer c.Close()

	if readTimeout != 0 {
		l.Debugw("set timeout", "readTimeout", readTimeout)
		c.SetReadDeadline(time.Now().Add(readTimeout))
	}

	bytesToWrite := connReq.getBytes()

	c.Write(bytesToWrite)

	buffRead := make([]byte, 32)

	count, err := c.Read(buffRead)
	if err != nil {
		l.Warnw("error reading", "err", err.Error())
		return connectResponse{}, err
	}
	buffRead = buffRead[:count]

	connResp := newConnectResponse(buffRead)
	if !connResp.matchesWithReq(connReq) {
		return connectResponse{}, err
	}
	return connResp, nil

}

func (impl *tracker) announce(ctx context.Context, u *url.URL, readTimeout time.Duration) (AnnounceResponse, error) {
	l := logger.Ctx(l_tracker, ctx).Sugar()
	l.Debugw("announce on", "host", u.Host)
	c, err := net.Dial("udp", u.Host)
	if err != nil {
		l.Error("error announce", "err", err.Error())
		return AnnounceResponse{}, err
	}
	defer c.Close()

	if readTimeout != 0 {
		l.Debugw("set announce deadline", "readTimeout", readTimeout)
		c.SetReadDeadline(time.Now().Add(readTimeout))
	}

	announceReq := newAnnounceRequest()

	announceReq.ConnID = impl.connID.id
	announceReq.TransactionID = newTransactionID()
	announceReq.InfoHash = impl.infoHash
	announceReq.PeerID = "-GO0000-0257f4bc7fa1"
	announceReq.Downloaded = uint64(impl.trackerInfo.Downloaded)
	announceReq.Uploaded = uint64(impl.trackerInfo.Uploaded)
	announceReq.Left = uint64(impl.trackerInfo.Left)
	announceReq.Port = impl.trackerInfo.Port

	l.Debugw("send announce", "announce", announceReq)

	c.Write(announceReq.getBytes())

	buffRead := make([]byte, 4096)

	n, err := c.Read(buffRead)
	if err != nil {
		l.Error("err reading announce response", "err", err.Error())
		return AnnounceResponse{}, err
	}
	buffRead = buffRead[:n]

	announceResp := newAnnounceResponse(buffRead)

	/* example resp
	* {Action:1 TxnID:2340680293 Interval:1697 Leechers:8 Seeders:0 Hosts: [...]
	 */
	//fmt.Printf("%+v", announceResp)
	return announceResp, nil

}

func newTransactionID() uint32 {
	transactionId := uint32(time.Now().UnixNano() ^ 0xdeadbeef)
	return transactionId

}
