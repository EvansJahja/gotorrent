package domain

import "errors"

type PieceList []byte

func NewPieceList(piecesCount int) PieceList {
	return make(PieceList, (piecesCount+7)/8)
}

func (p PieceList) ContainPiece(pieceNo uint32) bool {
	for i, b := range p {
		for j := 0; j < 8; j++ {
			key := uint32(i*8 + (7 - j))
			val := (b >> j & 1) == 1
			if key == pieceNo {
				return val
			}
		}
	}
	return false
}

func (p PieceList) SetPiece(pieceNo uint32) error {
	for i := range p {
		for j := 0; j < 8; j++ {
			key := uint32(i*8 + (7 - j))
			if key == pieceNo {
				p[i] |= 1 << j
				return nil
			}
		}
	}
	return errors.New("out of bound")
}
