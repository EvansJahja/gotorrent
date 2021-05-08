package runner

import (
	"fmt"
	"net/url"

	"example.com/gotorrent/lib/core/domain"
	"example.com/gotorrent/lib/files"
	"example.com/gotorrent/lib/peer"
	"example.com/gotorrent/lib/platform/udptracker"
)

type Runner struct {
	BasePath string
	Torrent  *domain.Torrent
	Files    files.Files
}

func (r Runner) Start() {

	for i, h := range r.Torrent.Hosts {
		if i%1 == 0 {
			go func(h domain.Host) {
				p := peer.NewPeer(h, r.Torrent)
				err := p.Connect()
				if err != nil {
					//
				} else {
					for {
						select {

						case n := <-p.GetNotification():
							switch n {
							case peer.NotiPiecesUpdated:
								fmt.Print("Noti pieces updated\n")
								v := p.GetHavePieces()
								if _, ok := v[0]; ok {
									fmt.Printf("piece 0 available\n")
								}
							case peer.NotiUnchocked:
								fmt.Print("Noti unchocked\n")
								p.RequestPiece(0)

							}
						}
					}

				}
			}(h)

		}
	}

}

func (r Runner) SetupTracker() {
	l := "***REMOVED***"
	u, _ := url.Parse(l)
	v := u.Query()
	fmt.Print(v)

	for i, ann := range v["tr"] {
		if i == 5 {
			r.connect(ann)
		}
	}
	fmt.Print("done")
}

func (r Runner) connect(announceUrl string) {
	u, _ := url.Parse(announceUrl)
	fmt.Print(u)

	announceResp, err := udptracker.Announce(u)
	if err != nil {
		return
	}

	r.Torrent.Lock()
	defer r.Torrent.Unlock()
	r.Torrent.Hosts = announceResp.Hosts
	r.Torrent.HintUpdated()

}
