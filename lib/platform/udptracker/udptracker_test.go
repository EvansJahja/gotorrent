package udptracker

import (
	"fmt"
	"net/url"
	"sync"
	"testing"
	"time"

	"bou.ke/monkey"
	"golang.org/x/net/nettest"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_start(t *testing.T) {

	tracker, _ := url.Parse("udp://127.0.0.1/announce")
	trackers := []*url.URL{tracker}
	u := UdpPeerList{
		InfoHash: []byte{1, 2, 3},
		Trackers: trackers,
	}
	u.Start()
	u.Stop()
}

func Test_connect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	conn, _ := nettest.NewLocalPacketListener("udp")

	fmt.Print(conn.LocalAddr().String())
	t1 := time.Date(2020, time.February, 12, 17, 18, 10, 0, time.UTC)
	t2 := t1.Add(5 * time.Second)
	t3 := t1.Add(10 * time.Second)
	t4 := t1.Add(15 * time.Second)
	t5 := t1.Add(20 * time.Second)
	times := []time.Time{t1, t2, t3, t4, t5}
	i := 0
	monkey.Patch(time.Now, func() time.Time {
		mockTime := times[i]
		fmt.Println(i)
		if i+1 < len(times) {
			i += 1
		}
		return mockTime

	})

	tracker, _ := url.Parse(fmt.Sprintf("udp://%s/announce", conn.LocalAddr().String()))
	trackers := []*url.URL{tracker}
	u := UdpPeerList{
		InfoHash: []byte{1, 2, 3},
		Trackers: trackers,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		b := make([]byte, 2048)
		n, respAddr, _ := conn.ReadFrom(b)
		b = b[:n]
		gotReq := newConnectRequestFromBytes(b)
		assert.Equal(t, uint32(0), gotReq.action)
		assert.Equal(t, 12345, gotReq.transactionId)
		connId := uint64(123)
		resp := connectResponse{
			action:        0,
			transactionId: 12345,
			connID:        connId,
		}
		conn.WriteTo(resp.getBytes(), respAddr)
		wg.Done()
	}()

	_, err := u.connect(tracker, 12345)
	assert.Error(t, err)
	wg.Wait()

}

func Test_getBytes(t *testing.T) {

	u := newConnectRequest()
	u.transactionId = 0xdeadbeef
	assert.Equal(t, true, len(u.getBytes()) > 0)

}
