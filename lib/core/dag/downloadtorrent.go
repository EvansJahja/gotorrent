package dag

import (
	"fmt"
	"io"

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
