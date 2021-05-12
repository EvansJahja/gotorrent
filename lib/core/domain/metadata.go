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

func (t Torrent) PiecesCount() int {
	return len(t.Pieces) / 20
}

func (t Torrent) TorrentLength() int {
	if t.Length != 0 {
		return t.Length
	}
	var l int
	for _, fp := range t.Files {
		l += fp.Length
	}
	return l
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
