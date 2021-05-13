package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewPieceList(t *testing.T) {
	p := NewPieceList(10)
	assert.False(t, p.ContainPiece(0))
	assert.NoError(t, p.SetPiece(0))
	assert.True(t, p.ContainPiece(0))

	assert.False(t, p.ContainPiece(9))
	assert.NoError(t, p.SetPiece(9))
	assert.True(t, p.ContainPiece(0))

	assert.NoError(t, p.SetPiece(15))
	assert.Error(t, p.SetPiece(16))
	assert.Error(t, p.SetPiece(17))

}
func Test_PieceList(t *testing.T) {
	p := PieceList([]byte{1<<0 | 1<<1 | 1<<3, 0})
	assert.True(t, p.ContainPiece(0))
	assert.True(t, p.ContainPiece(1))
	assert.False(t, p.ContainPiece(2))
	assert.True(t, p.ContainPiece(3))

	assert.False(t, p.ContainPiece(8))
	p.SetPiece(8)
	assert.True(t, p.ContainPiece(8))

	assert.False(t, p.ContainPiece(9))
	p.SetPiece(9)
	assert.True(t, p.ContainPiece(9))

}
