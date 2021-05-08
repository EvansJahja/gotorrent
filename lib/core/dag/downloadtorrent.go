package dag

import (
	"errors"
	"fmt"
	"io"
	"time"

	"example.com/gotorrent/lib/core/adapter/peer"

	"example.com/gotorrent/lib/core/domain"
)

// Runner  run from  iterator, send back data until said iterator is closed
// Task produces subtasks for runner to run, retrieve data from runner,
// and finally close the iterator

// DownloadStuffs:
// 1. Get list of things to download.
// 2. Download for each items we want to download
// 3. close iterator
type Task interface {
	Run() chan Task
}

type GetListOfDownloads struct {
	Globals Globals
	Resp    chan []domain.File
}

func (d GetListOfDownloads) Run() chan Task {
	task := make(chan Task)
	go func() {
		defer close(task)
		files := d.Globals.LocalRepo.GetFiles()
		d.Resp <- files
	}()
	return task
}

type DownloadStuffs struct {
	Globals Globals
}

func (d DownloadStuffs) Run() chan Task {
	task := make(chan Task)
	go func() {
		defer close(task)

		l := GetListOfDownloads{
			Globals: d.Globals,
			Resp:    make(chan []domain.File),
		}
		task <- l
		resp := <-l.Resp
		for i, f := range resp {
			if i > 3 {
				// TODO
				// Temp for easy debug
				break
			}
			task <- CompleteSingleFile{
				Globals: d.Globals,
				File:    f,
			}
		}

	}()

	return task

}

type CompleteSingleFile struct {
	Globals Globals
	File    domain.File
}

func (d CompleteSingleFile) Run() chan Task {
	task := make(chan Task)
	go func() {
		defer close(task)
		fmt.Printf("downloading %s\n", d.File)
		dp := DownloadPieces{
			Globals: d.Globals,
			Pieces:  []int{1, 3, 4},
			Resp:    make(chan DownloadPieceResp),
		}
		task <- dp
		<-dp.Resp
		fmt.Printf("download %s complete\n", d.File)

	}()
	return task
}
func (d CompleteSingleFile) String() string {
	return fmt.Sprintf("CompleteSingleFile")
	//return fmt.Sprintf("CompleteSingleFile{%s}", d.File.Path)
}

type DownloadPieces struct {
	Globals Globals
	Pieces  []int
	Resp    chan DownloadPieceResp
}

type DownloadPieceResp struct {
	Piece  int
	Reader io.Reader
}

func (d DownloadPieces) Run() chan Task {
	tasks := make(chan Task)
	go func() {
		defer close(tasks)
		conn := GetPeers{Globals: d.Globals, Resp: make(chan []peer.Peer)}
		tasks <- conn
		peers := <-conn.Resp
		for _, p := range peers {
			havePieces := p.GetHavePieces()
			for _, wantPiece := range d.Pieces {
				if _, ok := havePieces[wantPiece]; ok {
					// peer has piece
					reader := p.RequestPiece(wantPiece)
					d.Resp <- DownloadPieceResp{
						Piece:  wantPiece,
						Reader: reader,
					}

				}
			}
		}
		defer close(d.Resp)

	}()
	return tasks
}

type GetHosts struct {
	Globals Globals
	Resp    chan []domain.Host
}

func (d GetHosts) Run() chan Task {
	tasks := make(chan Task)
	go func() {
		defer close(tasks)
		var hosts []domain.Host

		cache := GetHostsFromCache{Globals: d.Globals, Resp: make(chan []domain.Host)}
		tasks <- cache
		hosts = <-cache.Resp
		if len(hosts) == 0 {
			announce := GetHostsFromAnnounce{Globals: d.Globals, Resp: make(chan []domain.Host)}
			tasks <- announce
			hosts = <-announce.Resp

			s := SetHostsToCache{
				Globals: d.Globals,
				Hosts:   hosts,
			}
			tasks <- s
		}
		d.Resp <- hosts

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
		var peers []peer.Peer

		type key string
		kPeers := key("peers")

		v, err := d.Globals.Cache.Cached(kPeers, func() (interface{}, error) {
			getHosts := GetHosts{Globals: d.Globals, Resp: make(chan []domain.Host)}
			tasks <- getHosts
			hosts := <-getHosts.Resp

			for _, h := range hosts {
				p := d.Globals.PeerFactory.New(h)
				peers = append(peers, p)
			}
			return peers, nil
		})
		if err != nil {
			panic(err)
		}
		peers = v.([]peer.Peer)

		d.Resp <- peers
	}()
	return tasks
}

type HostsWithTimestamp struct {
	Hosts []domain.Host
	Time  time.Time
}
type GetHostsFromCache struct {
	Globals Globals
	Resp    chan []domain.Host
}

func (d GetHostsFromCache) Run() chan Task {
	tasks := make(chan Task)
	go func() {
		defer close(tasks)
		kHosts := "hosts"
		type key string
		kHostsCKey := key("hosts")
		v, err := d.Globals.Cache.Cached(kHostsCKey, func() (interface{}, error) {

			var persistHost HostsWithTimestamp

			var hosts []domain.Host
			if err := d.Globals.PersistentMetadata.Get(kHosts, &persistHost); err == nil {
				t := time.Now().Add(30 * time.Minute)
				if persistHost.Time.Before(t) {
					hosts = persistHost.Hosts
					fmt.Println("Using persisted host")
				}
			}
			return hosts, nil
		})
		if err != nil {
			panic(err)
		}

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
		kHosts := "hosts"
		persistHost := HostsWithTimestamp{
			Hosts: d.Hosts,
			Time:  time.Now(),
		}
		d.Globals.PersistentMetadata.Put(kHosts, persistHost)
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
		type key string
		kAnnounce := key("announce")
		v, err := d.Globals.Cache.Cached(kAnnounce, func() (interface{}, error) {
			fmt.Println("GET PEERS FROM ANNOUNCE")
			hosts := d.Globals.PeerList.GetPeers()
			if len(hosts) == 0 {
				return nil, errors.New("no hosts")
			}
			return hosts, nil

		})
		if err != nil {
			panic("uh oh")
		}
		hosts := v.([]domain.Host)
		d.Resp <- hosts
	}()
	return tasks
}

/*

type RunMeOnce struct {
	Resp    chan string
	Globals Globals
}

func (d RunMeOnce) Run() chan Task {
	task := make(chan Task)
	go func() {
		defer close(task)

		type key string
		kMutex := key("mutex")
		kData := key("data")
		var mut *sync.Mutex
		if v, ok := d.Globals.ValueCache.Get(kMutex); ok {
			mut = v.(*sync.Mutex)
		} else {
			mut = &sync.Mutex{}
			d.Globals.ValueCache.Set(kMutex, mut)
		}

		var resp string

		mut.Lock()
		if v, ok := d.Globals.ValueCache.Get(kData); ok {
			resp = v.(string)
			goto Done
		}

		fmt.Printf("RUN ME ONCE\n")
		time.Sleep(5 * time.Second)
		resp = "RunMeOnceResponse"
		d.Globals.ValueCache.Set(kData, resp)

	Done:
		mut.Unlock()

		d.Resp <- resp

	}()
	return task
}
*/

/*
type Node interface {
	Start() []Node
	Stop()
}

type DownloadTorrent struct {
	Globals Globals
}

func (d DownloadTorrent) Start() []Node {
	p := FetchToFileSystem{
		Globals: d.Globals,
	}
	return []Node{p}
}
func (d DownloadTorrent) Stop() {
}

type FetchToFileSystem struct {
	Globals Globals
}

func (d FetchToFileSystem) Start() []Node {
	var deps []Node
	for _, f := range d.Globals.LocalRepo.GetFiles() {
		d := DownloadFile{
			Globals: d.Globals,
			File:    f,
		}
		d.Globals.Runner(d)
		deps = append(deps, d)
	}
	return deps

}

func (d FetchToFileSystem) Stop() {
}

type DownloadFile struct {
	Globals Globals
	File    domain.File
}

func (d DownloadFile) String() string {
	return fmt.Sprintf("DownloadFile{%s}", d.File.Path)
}

func (d DownloadFile) Start() []Node {
	return []Node{}

}

func (d DownloadFile) Stop() {

}

*/
