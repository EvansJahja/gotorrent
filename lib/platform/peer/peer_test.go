package peer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_handshake(t *testing.T) {

	handshakeReq := handshake{
		proto:        protoBitTorrent,
		featureFlags: 0x10_00_00_00_00_10_00_11,
		infoHash:     "***REMOVED***",
		peerID:       []byte("-GO0000-0257f4bc7fa1"),
	}

	b := handshakeReq.getBytes()
	handshakeReconstruct := newHandshake(b)

	assert.Equal(t, handshakeReq, handshakeReconstruct)

}
