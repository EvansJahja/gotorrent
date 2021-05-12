package main

import (
	"fmt"
	"net/url"

	"example.com/gotorrent/lib/core/domain"
	"example.com/gotorrent/lib/platform/udptracker"
)

func main() {
	magnetURI := "***REMOVED***"
	//magnetURI := os.Args[1]

	u, _ := url.Parse(magnetURI)
	magnetU := domain.Magnet{Url: u}
	trackers := magnetU.Trackers()
	infoHash := magnetU.InfoHash()

	udpPeerList := udptracker.UdpPeerList{
		InfoHash: infoHash,
		Trackers: trackers,
	}

	for _, trackerU := range trackers {
		resp, err := udpPeerList.Announce(trackerU)
		if err != nil {
			// fmt.Printf("error: %s\n", err.Error())
			//os.Exit(1)
			continue
		}
		fmt.Printf("%+v\n", resp)
	}

}
