package peer

import (
	"bytes"
	"encoding/binary"
)

type handshake struct {
	proto        string
	featureFlags uint64
	infoHash     []byte
	peerID       []byte
}

func (u handshake) matches(v handshake) bool {
	return u.proto == v.proto && bytes.Equal(u.infoHash, v.infoHash)
}

func (h handshake) getBytes() []byte {
	handshakeBytes := make([]byte, 1024)

	protoLen := len(h.proto)

	n := 0
	n += copy(handshakeBytes[n:], []byte{byte(protoLen)})
	n += copy(handshakeBytes[n:], []byte(h.proto))

	binary.BigEndian.PutUint64(handshakeBytes[n:], h.featureFlags)
	n += 8

	n += copy(handshakeBytes[n:], h.infoHash)
	n += copy(handshakeBytes[n:], []byte(h.peerID))

	handshakeBytes = handshakeBytes[:n]

	return handshakeBytes

}
func newHandshake(b []byte) handshake {
	var h handshake

	protoLen := b[0]
	n := 1

	if int(protoLen) != len(protoBitTorrent) {
		return h
	}

	h.proto = string(b[n : n+int(protoLen)])
	n += len(h.proto)
	h.featureFlags = binary.BigEndian.Uint64(b[n:])
	n += 8

	h.infoHash = b[n : n+20]
	n += 20

	h.peerID = b[n : n+20]

	return h

}
