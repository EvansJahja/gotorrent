package domain

import (
	"bytes"
	"crypto/sha1"

	"github.com/jackpal/bencode-go"
)

type Metadata []byte

type Torrent struct {
	Name        string
	PieceLength int `bencode:"piece length"`
	Pieces      string

	Files  []FileInfo
	Length int
	Path   []string
}

type FileInfo struct {
	Length int
	Path   []string
}

func (m Metadata) InfoHash() []byte {

	s := sha1.New()
	s.Write(m)
	b := s.Sum(nil)

	return b
}
func (m Metadata) MustParse() Torrent {
	var t Torrent

	reader := bytes.NewReader(m)

	err := bencode.Unmarshal(reader, &t)
	if err != nil {
		panic(err)
	}

	return t

}
