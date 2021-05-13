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
	p := NewPieceList(10)
	p.SetPiece(0)
	p.SetPiece(1)
	p.SetPiece(3)
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
func Test_ContainPiece(t *testing.T) {
	p := PieceList([]byte{0b10110100, 0b11110000})
	assert.True(t, p.ContainPiece(0))
	assert.False(t, p.ContainPiece(1))
	assert.True(t, p.ContainPiece(2))
	assert.True(t, p.ContainPiece(3))
	assert.False(t, p.ContainPiece(4))
	assert.True(t, p.ContainPiece(5))
	assert.False(t, p.ContainPiece(6))
	assert.False(t, p.ContainPiece(7))

	assert.True(t, p.ContainPiece(8))
	assert.True(t, p.ContainPiece(9))
	assert.True(t, p.ContainPiece(10))
	assert.True(t, p.ContainPiece(11))
	assert.False(t, p.ContainPiece(12))
	assert.False(t, p.ContainPiece(13))
	assert.False(t, p.ContainPiece(14))
	assert.False(t, p.ContainPiece(15))

}
