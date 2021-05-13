//go:generate mockgen -destination ../../mocks/net/net.go net Conn
package peer

import (
	"encoding/hex"
	"testing"

	"example.com/gotorrent/lib/core/adapter/peer"
	"example.com/gotorrent/lib/core/domain"
	mock_net "example.com/gotorrent/lib/mocks/net"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {
	t.Run("handleMessage should send piece request channel when receive Request Piece", func(t *testing.T) {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conn := mock_net.NewMockConn(ctrl)

		p := new(domain.Host{}, []byte{}).(*peerImpl)
		p.conn = conn

		conn.EXPECT().Write(gomock.Any()).DoAndReturn(func(b []byte) (int, error) {
			p.handleMessage(b[4:]) // strip length
			return len(b), nil
		})
		p.RequestPiece(4, 5, 8) // not tested, only to get []byte in expect above

		req := <-p.PieceRequests()
		expReq := peer.PieceRequest{
			PieceNo: 4,
			Begin:   5,
			Length:  8,
		}
		assert.Equal(t, expReq.PieceNo, req.PieceNo)
		assert.Equal(t, expReq.Begin, req.Begin)
		assert.Equal(t, expReq.Length, req.Length)
	})
}

func Test_handshake(t *testing.T) {
	infoHash, _ := hex.DecodeString("a4ef8a65e78a69eedf588cb87e382d382a37baab")

	handshakeReq := handshake{
		proto:        protoBitTorrent,
		featureFlags: 0x10_00_00_00_00_10_00_11,
		infoHash:     infoHash,
		peerID:       []byte("-GO0000-0257f4bc7fa1"),
	}

	b := handshakeReq.getBytes()
	handshakeReconstruct := newHandshake(b)

	assert.Equal(t, handshakeReq, handshakeReconstruct)

}
