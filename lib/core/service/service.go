package service

import (
	"time"

	"example.com/gotorrent/lib/core/adapter/peerlist"
	"example.com/gotorrent/lib/core/domain"
)

type TorrentService interface {
	Start() <-chan struct{}
}

type TorrentImpl struct {
	MagnetURI  string
	TargetPath string
	PeerList   peerlist.PeerRepo

	doneCh       chan struct{}
	hosts        []domain.Host
	peerServices []PeerService
}

var _ TorrentService = TorrentImpl{}

func (t TorrentImpl) Start() <-chan struct{} {
	t.doneCh = make(chan struct{})
	go func() {
		t.run()
	}()
	return t.doneCh
}

func (t TorrentImpl) run() {
	t.hosts = t.PeerList.GetPeers()
	for _, host := range t.hosts {
		peerSvc := PeerServiceImpl{Host: host}
		t.peerServices = append(t.peerServices, peerSvc)
	}

	for _, peerSvc := range t.peerServices {
		peerSvc.Start()
	}

	time.Sleep(1 * time.Second)
	t.doneCh <- struct{}{}
}
