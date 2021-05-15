package echohttp

import (
	"errors"
	"fmt"
	"strconv"

	"example.com/gotorrent/lib/core/adapter/peer"
	"example.com/gotorrent/lib/core/service/peerpool"
	"github.com/labstack/echo/v4"
)

type HTTPServe struct {
	PeerPool peerpool.PeerPool
}

func (h HTTPServe) Start() {
	go func() {

		e := echo.New()
		e.Add("GET", "/health", h.health)
		e.Add("GET", "/peers", h.peers)
		e.Add("GET", "/peer/:hashID", h.peerID)

		e.Start(":8080")
	}()

}

func (h HTTPServe) health(c echo.Context) error {
	return c.JSON(200, "OK")
}
func (h HTTPServe) peers(c echo.Context) error {

	var filterList []peerpool.PeerFilter
	v := c.QueryParams()
	for _, f := range v["filter"] {
		switch f {
		case "download":
			filterList = append(filterList, peerpool.FilterHasDownload)
		case "upload":
			filterList = append(filterList, peerpool.FilterHasUpload)
		case "downloading":
			filterList = append(filterList, peerpool.FilterIsDownloading)
		case "uploading":
			filterList = append(filterList, peerpool.FilterIsUploading)
		case "choked":
			filterList = append(filterList, peerpool.FilterChoking)
		case "unchoked":
			filterList = append(filterList, peerpool.FilterNotChoking)
		}
	}
	if pieceNoStr := c.QueryParam("piece"); pieceNoStr != "" {
		pieceNo, err := strconv.Atoi(pieceNoStr)
		if err == nil {
			filterList = append(filterList, peerpool.FilterHasPiece(uint32(pieceNo)))
		}
	}

	peers := h.PeerPool.Peers(append(filterList, peerpool.FilterConnected)...)

	type peerStatType struct {
		Hostname      string
		HashedPeerID  string
		DownloadRate  string
		UploadRate    string
		TotalDownload string
		TotalUpload   string
	}
	peerStats := make([]peerStatType, 0, len(peers))

	for _, peerObj := range peers {
		s := peerStatType{
			Hostname:      peerObj.Hostname(),
			HashedPeerID:  peerObj.GetID(),
			DownloadRate:  fmt.Sprintf("%f kBps", peerObj.GetDownloadRate()),
			UploadRate:    fmt.Sprintf("%f kBps", peerObj.GetUploadRate()),
			TotalDownload: fmt.Sprintf("%f kB", float64(peerObj.GetDownloadBytes())/1000),
			TotalUpload:   fmt.Sprintf("%f kB", float64(peerObj.GetUploadBytes())/1000),
		}
		peerStats = append(peerStats, s)
	}
	return c.JSON(200, peerStats)

}

func (h HTTPServe) peerID(c echo.Context) error {
	// find peer
	hashID := c.Param("hashID")
	p, err := h.findPeerByHash(hashID)
	if err != nil {
		return c.JSON(404, "not found")
	}
	type peerDetailStatType struct {
		Hostname      string
		DownloadRate  string
		UploadRate    string
		TotalDownload string
		TotalUpload   string
		Pieces        []uint32
	}
	var pieces []uint32
	theirPieces := p.TheirPieces()
	for i := uint32(0); i < theirPieces.ApproxPieceCount(); i++ {
		if theirPieces.ContainPiece(i) {
			pieces = append(pieces, i)
		}
	}

	peerDetail := peerDetailStatType{
		Hostname: p.Hostname(),

		DownloadRate:  fmt.Sprintf("%f kBps", p.GetDownloadRate()),
		UploadRate:    fmt.Sprintf("%f kBps", p.GetUploadRate()),
		TotalDownload: fmt.Sprintf("%f kB", float64(p.GetDownloadBytes())/1000),
		TotalUpload:   fmt.Sprintf("%f kB", float64(p.GetUploadBytes())/1000),

		Pieces: pieces,
	}
	return c.JSON(200, peerDetail)

}

func (h HTTPServe) findPeerByHash(hashID string) (peer.Peer, error) {

	peers := h.PeerPool.Peers(peerpool.FilterConnected)
	for _, peerObj := range peers {
		if peerObj.GetID() == hashID {
			return peerObj, nil
		}
	}
	return nil, errors.New("not found")
}
