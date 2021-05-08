package extensions

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"

	"github.com/jackpal/bencode-go"
)

type ExtHandler interface {
	Init()
	HandleCommand(msgId byte, msgVal []byte)
	FetchMetadata() <-chan []byte
}

type extHandler struct {
	conn             net.Conn
	peerExtendedInfo map[string]interface{}
	ourExtendedInfo  map[string]interface{}

	sendMsgFn     SendMsgFn
	extensionList extensionList
}

var _ ExtHandler = &extHandler{}

func NewExtHandler(conn net.Conn, sendMsgFn SendMsgFn) ExtHandler {
	e := extHandler{
		conn:      conn,
		sendMsgFn: sendMsgFn,
	}
	return &e
}

type extensionList struct {
	utMetadata *utMetadata
}

type SendMsgFn func(msg []byte, cmd byte)

type SendCmdFn func(msg []byte)

type A interface {
	Startup()
	TearDown()
	HandleCommand(msgId byte, msgVal []byte)
	SetRxOn(msgId byte)
	SetTxOn(msgId byte)
}

func (h *extHandler) Init() {
	utMetadata := utMetadata{
		SendMsgFn: h.sendMsg,
	}

	utMetadata.SetRxOn(9)
	h.extensionList.utMetadata = &utMetadata

	// Send our own handshake
	h.ourExtendedInfo = map[string]interface{}{
		"m": map[string]interface{}{
			"ut_metadata": 9,
		},
		"metadata_size": 0,
	}
}

func (h *extHandler) SetExtendedInfo(m map[string]interface{}) {
	h.peerExtendedInfo = m
	b, err := digForM(h.peerExtendedInfo, "ut_metadata")
	if err != nil {
		return
	}
	h.extensionList.utMetadata.SetTxOn(b)
}

func (h *extHandler) Startup() {
	h.startup()

	h.extensionList.utMetadata.Startup()
}

func (h *extHandler) FetchMetadata() <-chan []byte {
	return h.extensionList.utMetadata.RequestMetadata()
}

func (h *extHandler) TearDown() {
	h.extensionList.utMetadata.TearDown()
}

func (h *extHandler) HandleCommand(msgId byte, msgVal []byte) {
	switch msgId {
	case 0: // Handshake
		data, err := bencode.Decode(bufio.NewReader(bytes.NewBuffer(msgVal)))
		if err != nil {
			fmt.Print(err)
			return
		}
		if dataMap, ok := data.(map[string]interface{}); ok {
			h.SetExtendedInfo(dataMap)
			h.Startup()
		}
	default:
		h.extensionList.utMetadata.HandleCommand(msgId, msgVal)
	}
}

func (h *extHandler) sendMsg(msg []byte) {
	h.sendMsgFn(msg, 20)
}

func (h *extHandler) startup() {
	var bbuf bytes.Buffer

	err := bencode.Marshal(bufio.NewWriter(&bbuf), h.ourExtendedInfo)

	b := bbuf.Bytes()
	if len(b) == 0 {
		// Empty extended info
		return
	}

	if err != nil {
		fmt.Print(err)
	}

	buf := make([]byte, len(b)+1)
	copy(buf, []byte{0x00})
	copy(buf[1:], b)

	h.sendMsg(buf)

}

func digForM(m map[string]interface{}, target string) (byte, error) {
	if mInterface, ok := m["m"]; ok {
		if mMap, ok := mInterface.(map[string]interface{}); ok {
			if vCmd, ok := mMap[target]; ok {
				if iCmd, ok := vCmd.(int64); ok {
					res := byte(iCmd)
					return res, nil
				}
			}
		}
	}

	return 0, errors.New("not found")
}
