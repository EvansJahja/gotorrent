package udptracker

/*
 * BEP 15 UDPTracker
 */
import (
	"net/url"

	"example.com/gotorrent/lib/core/adapter/peerlist"
	"example.com/gotorrent/lib/core/domain"
	"example.com/gotorrent/lib/logger"
)

// UDP peer list handle multiple trackers
type UdpPeerList struct {
	InfoHash     []byte
	Trackers     []*url.URL // TODO: rename to tracker urls
	trackerImpls []*tracker
	hosts        []domain.Host
	trackerInfo  TrackerInfo
}

var _ peerlist.PeerRepo = &UdpPeerList{}

var l_udptracker = logger.Named("udptracker")

type TrackerInfo struct {
	Uploaded   int
	Downloaded int
	Left       int
	Port       uint16
}

func (peerList *UdpPeerList) SetInfo(newInfo TrackerInfo) {
	peerList.trackerInfo = newInfo
	for _, trackerImpl := range peerList.trackerImpls {
		trackerImpl.updateInfo(newInfo)
	}
}

func (peerList *UdpPeerList) Start() {
	l_udptracker.Debug("starting udptracker")

	for _, trackerUrl := range peerList.Trackers {
		hostChan := make(chan domain.Host, 100)

		trackerImpl := &tracker{
			trackerUrl:  trackerUrl,
			newHostChan: hostChan,
			infoHash:    peerList.InfoHash,
		}

		trackerImpl.updateInfo(peerList.trackerInfo)
		peerList.trackerImpls = append(peerList.trackerImpls, trackerImpl)
		l_udptracker.Sugar().Debugw("running trackerImpl", "url", trackerUrl)
		//l.Info("running trackerImpl for ")
		go trackerImpl.run()
		go func() {
			for h := range hostChan {
				peerList.addHost(h)
			}
		}()
	}
}
func (peerList *UdpPeerList) addHost(newHost domain.Host) {
	// Check if host exist
	for _, h := range peerList.hosts {
		if h.Equal(newHost) {
			return
		}
	}
	// this is new
	peerList.hosts = append(peerList.hosts, newHost)
}

func (peerList *UdpPeerList) GetPeers() []domain.Host {
	return peerList.hosts
}
