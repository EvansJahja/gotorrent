package udptracker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getBytes(t *testing.T) {

	u := newConnectRequest()
	u.transactionId = 0xdeadbeef
	assert.Equal(t, true, len(u.getBytes()) > 0)

}
