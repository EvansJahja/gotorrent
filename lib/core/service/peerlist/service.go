package peerlist

import (
	"errors"
	"fmt"
	"time"

	"example.com/gotorrent/lib/core/adapter/cache"
	"example.com/gotorrent/lib/core/adapter/peerlist"
	"example.com/gotorrent/lib/core/adapter/persistentmetadata"
	"example.com/gotorrent/lib/core/domain"
)

type Service interface {
	GetHosts() ([]domain.Host, error)
}

type Impl struct {
	Cache cache.Cache
	//PeerFactory        peer.PeerFactory
	PersistentMetadata persistentmetadata.PersistentMetadata
	PeerList           peerlist.PeerRepo
}

var _ Service = Impl{}

type hostsWithTimestamp struct {
	Hosts []domain.Host
	Time  time.Time
}

// Unused. this service should only give hosts
/*
func (impl Impl) GetPeers() ([]peer.Peer, error) {
	var peers []peer.Peer

	type key string
	kPeers := key("peers")

	v, err := impl.Cache.Cached(kPeers, func() (interface{}, error) {
		hosts, err := impl.GetHosts()
		if err != nil {
			return nil, err
		}

		for _, h := range hosts {
			p := impl.PeerFactory.New(h)
			peers = append(peers, p)
		}
		return peers, nil
	})
	if err != nil {
		return nil, err
	}
	peers = v.([]peer.Peer)

	return peers, nil

}
*/

func (impl Impl) GetHosts() ([]domain.Host, error) {
	var hosts []domain.Host
	var err error

	hosts, err = impl.getHostsFromCache()
	if err != nil {
		hosts, err = impl.getHostsFromAnnounce()
		if err != nil {
			return nil, err
		}
		_ = impl.setHostsToCache(hosts)
	}
	return hosts, nil

}

func (impl Impl) getHostsFromCache() ([]domain.Host, error) {

	kHosts := "hosts"
	type key string
	kHostsCKey := key("hosts")
	v, err := impl.Cache.Cached(kHostsCKey, func() (interface{}, error) {

		var persistHost hostsWithTimestamp

		var hosts []domain.Host
		if err := impl.PersistentMetadata.Get(kHosts, &persistHost); err == nil {
			cacheExpTime := persistHost.Time.Add(30 * time.Minute)
			fmt.Printf("Cache expires in %s\n", cacheExpTime.Format(time.RFC3339))
			if time.Now().Before(cacheExpTime) {
				hosts = persistHost.Hosts
				fmt.Println("Using persisted host")
			} else {
				return nil, errors.New("cache expired")
			}
		} else {
			return nil, err
		}
		return hosts, nil

	})
	if err != nil {
		return nil, err
	}
	hosts := v.([]domain.Host)
	return hosts, nil
}

func (impl Impl) getHostsFromAnnounce() ([]domain.Host, error) {
	type key string
	kAnnounce := key("announce")
	v, err := impl.Cache.Cached(kAnnounce, func() (interface{}, error) {
		fmt.Println("GET PEERS FROM ANNOUNCE")
		hosts := impl.PeerList.GetPeers()
		if len(hosts) == 0 {
			return nil, errors.New("no hosts")
		}
		return hosts, nil

	})
	if err != nil {
		return nil, err
	}
	hosts := v.([]domain.Host)
	return hosts, nil
}

func (impl Impl) setHostsToCache(hosts []domain.Host) error {
	kHosts := "hosts"
	persistHost := hostsWithTimestamp{
		Hosts: hosts,
		Time:  time.Now(),
	}
	return impl.PersistentMetadata.Put(kHosts, persistHost)
}

/*

type GetHosts struct {
	Globals Globals
	Resp    chan []domain.Host
}

func (d GetHosts) Run() chan Task {
	tasks := make(chan Task)
	go func() {
		defer close(tasks)
	}()
	return tasks
}

type GetPeers struct {
	Globals Globals
	Resp    chan []peer.Peer
}

func (d GetPeers) Run() chan Task {
	tasks := make(chan Task)
	go func() {
		defer close(tasks)
	}()
	return tasks
}

type GetHostsFromCache struct {
	Globals Globals
	Resp    chan []domain.Host
}

func (d GetHostsFromCache) Run() chan Task {
	tasks := make(chan Task)
	go func() {
		defer close(tasks)

		d.Resp <- v.([]domain.Host)

	}()
	return tasks
}

type SetHostsToCache struct {
	Globals Globals
	Hosts   []domain.Host
	Resp    chan struct{}
}

func (d SetHostsToCache) Run() chan Task {
	tasks := make(chan Task)
	go func() {
	}()
	return tasks
}

type GetHostsFromAnnounce struct {
	Globals Globals
	Resp    chan []domain.Host
}

func (d GetHostsFromAnnounce) Run() chan Task {
	tasks := make(chan Task)
	go func() {
		defer close(tasks)
	}()
	return tasks
}
*/
