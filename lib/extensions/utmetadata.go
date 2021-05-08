package extensions

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/jackpal/bencode-go"
)

type utMetadata struct {
	//extHandler ExtHandler
	SendMsgFn SendCmdFn

	TxMsgId byte
	RxMsgId byte

	metadataChan chan []byte
}

var _ A = &utMetadata{}

func (h utMetadata) Startup() {
}

func (h *utMetadata) RequestMetadata() <-chan []byte {
	h.metadataChan = make(chan []byte)
	go h.requestMetadata()

	return h.metadataChan
}

func (h *utMetadata) requestMetadata() {
	for h.TxMsgId == 0 {
		time.Sleep(1000 * time.Millisecond)
		go h.requestMetadata()
		return
	}
	/*
		if h.TxMsgId == 0 {
			return
		}
	*/

	d := map[string]interface{}{
		"msg_type": 0,
		"piece":    0,
	}

	var buf bytes.Buffer

	err := bencode.Marshal(&buf, d)
	if err != nil {
		fmt.Println(err)
		return
	}

	b, err := io.ReadAll(&buf)
	if err != nil {
		fmt.Println(err)
		return
	}

	sendBuf := make([]byte, len(b)+1)
	copy(sendBuf[0:], []byte{h.TxMsgId})
	copy(sendBuf[1:], b)

	h.SendMsgFn(sendBuf)
}

func (h utMetadata) TearDown() {

}
func (h utMetadata) HandleCommand(msgId byte, msgVal []byte) {
	if msgId != h.RxMsgId {
		return
	}

	di, err := bencode.Decode(bufio.NewReader(bytes.NewBuffer(msgVal)))
	if err != nil {
		return
	}
	d, ok := di.(map[string]interface{})
	if !ok {
		return
	}
	size, ok := d["total_size"].(int64)
	if !ok {
		return
	}
	fmt.Print(d)

	// Assume that the rest is dict size is len of msgVal - size.
	// Bad assumption I think?

	metaStart := len(msgVal) - int(size)

	metadata := msgVal[metaStart:]

	go func() {
		h.metadataChan <- metadata
		fmt.Println("done send channel")
	}()

}

func (h *utMetadata) SetRxOn(msgId byte) {
	h.RxMsgId = msgId
}

func (h *utMetadata) SetTxOn(msgId byte) {
	h.TxMsgId = msgId
}
