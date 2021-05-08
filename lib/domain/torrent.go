package domain

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/jackpal/bencode-go"
)

type Torrent struct {
	Name     string
	BasePath string
	InfoHash string
	Hosts    []Host
	Metadata []byte
	sync.RWMutex
	updatedChan chan struct{}

	metadata *Meta
}

func (t *Torrent) GetUpdatedChan() <-chan struct{} {
	if t.updatedChan == nil {
		t.updatedChan = make(chan struct{})
	}
	return t.updatedChan
}

func (t *Torrent) HintUpdated() {
	t.updatedChan <- struct{}{}
}

type Meta struct {
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

func (t *Torrent) GetMeta() Meta {
	t.RLock()
	defer t.RUnlock()

	if t.metadata != nil {
		return *t.metadata
	}

	var m Meta

	s := sha1.New()
	s.Write(t.Metadata)
	b := s.Sum(nil)
	if hex.EncodeToString(b) != t.InfoHash {
		fmt.Printf("Invalid sum")
		return Meta{}
	}

	reader := bytes.NewReader(t.Metadata)
	bencode.Unmarshal(reader, &m)
	t.metadata = &m
	return m

}
