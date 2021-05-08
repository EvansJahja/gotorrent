package service

import (
	"example.com/gotorrent/lib/core/domain"
)

type PeerService interface {
	Start()
}

type PeerServiceImpl struct {
	Host domain.Host
}

var _ PeerService = PeerServiceImpl{}

func (p PeerServiceImpl) Start() {

}
