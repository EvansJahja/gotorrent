package echohttp

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"example.com/gotorrent/lib/core/adapter/peer"
	"example.com/gotorrent/lib/core/bucketdownload"
	"example.com/gotorrent/lib/core/domain"
	"example.com/gotorrent/lib/core/service/peerpool"
	"example.com/gotorrent/lib/files"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nvellon/hal"
)

type HTTPServe struct {
	PeerPool  peerpool.PeerPool
	Bucket    bucketdownload.Bucket
	Files     *files.Files
	OurPieces *domain.PieceList
}

var peerRoute *echo.Route
var peersRoute *echo.Route
var fileRoute *echo.Route

func allows(s []string) func(c echo.Context) error {
	return func(c echo.Context) error {
		methods := strings.Join(s, ",")
		c.Response().Header().Set("Allow", methods)
		c.Response().WriteHeader(200)
		return nil

	}
}

func (h *HTTPServe) Start() {
	go func() {
		e := echo.New()
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     []string{"localhost"},
			AllowCredentials: true,
			AllowOriginFunc:  func(_ string) (bool, error) { return true, nil },
			ExposeHeaders:    []string{"Allow"},
		}))
		e.Add("GET", "/health", h.health)
		peersRoute = e.GET("/peers", h.peers)
		peerRoute = e.GET("/peer/:hashID", h.peerID)
		e.HEAD("/peer/:hashID", allows([]string{"get"}))
		e.GET("/downloads", h.downloads)
		e.GET("/files", h.files)
		fileRoute = e.GET("/file/:id", h.file)
		//e.GET("/", h.root)

		e.Start(":8080")
	}()

}
func (h *HTTPServe) files(c echo.Context) error {
	if h.Files == nil {
		return c.String(http.StatusServiceUnavailable, "files not available")
	}
	if h.OurPieces == nil {
		return c.String(http.StatusServiceUnavailable, "our pieces not available")
	}

	r := struct {
		CompletedFiles []string
	}{}
	resp := hal.NewResource(r, c.Request().RequestURI)
	type fileinfo struct {
		Path             string
		Pieces           []uint32
		OurHavePieces    []uint32
		OurNotHavePieces []uint32
	}

	for i, fp := range h.Files.Torrent.Files {
		fPath := h.Files.GetRelativePath(fp.Path)
		pieces := h.Files.GetPiecesFromFile(fPath)
		ourHavePiece := make([]uint32, 0, len(pieces))
		ourNotHavePiece := make([]uint32, 0, len(pieces))
		for _, pc := range pieces {
			if h.OurPieces.ContainPiece(pc) {
				ourHavePiece = append(ourHavePiece, pc)
			} else {
				ourNotHavePiece = append(ourNotHavePiece, pc)
			}
		}

		fileInfo := fileinfo{
			Path:             fPath,
			Pieces:           pieces,
			OurHavePieces:    ourHavePiece,
			OurNotHavePieces: ourNotHavePiece,
		}

		linkToFile := c.Echo().Reverse(fileRoute.Name, i)
		fileInfoResp := hal.NewResource(fileInfo, linkToFile)
		resp.Embedded.Add("file", fileInfoResp)
		resp.AddLink("file", hal.NewLink(linkToFile, hal.LinkAttr{"title": fPath}))
	}

	respB, _ := resp.MarshalJSON()
	return c.Blob(200, "application/hal+json", respB)
}

func (h *HTTPServe) file(c echo.Context) error {
	return nil
}

func (h *HTTPServe) downloads(c echo.Context) error {
	if h.Bucket == nil {
		return c.String(http.StatusServiceUnavailable, "bucket not available")
	}
	stats := h.Bucket.Stats()
	resp := hal.NewResource(stats, c.Request().RequestURI)

	respB, _ := resp.MarshalJSON()
	return c.Blob(200, "application/hal+json", respB)
}

func (h *HTTPServe) health(c echo.Context) error {
	return c.JSON(200, "OK")
}
func (h *HTTPServe) peers(c echo.Context) error {
	if h.PeerPool == nil {
		return c.String(http.StatusServiceUnavailable, "peerpool not available")
	}

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

	e := make(hal.Embedded)
	poolDownloadRate, poolUploadRate, poolDownloadTotal, poolUploadTotal := h.PeerPool.GetNetworkStats()

	for _, peerObj := range peers {
		s := peerStatType{
			Hostname:      peerObj.Hostname(),
			HashedPeerID:  peerObj.GetID(),
			DownloadRate:  fmt.Sprintf("%f kBps", peerObj.GetDownloadRate()),
			UploadRate:    fmt.Sprintf("%f kBps", peerObj.GetUploadRate()),
			TotalDownload: fmt.Sprintf("%f kB", float64(peerObj.GetDownloadBytes())/1000),
			TotalUpload:   fmt.Sprintf("%f kB", float64(peerObj.GetUploadBytes())/1000),
		}
		r := hal.NewResource(s, c.Echo().Reverse(peerRoute.Name, peerObj.GetID()))
		e.Add("peer", r)

	}

	type peerPoolStatType struct {
		NumOfPeers    int
		DownloadRate  string
		UploadRate    string
		TotalDownload string
		TotalUpload   string
	}
	peerPoolStat := peerPoolStatType{
		NumOfPeers:    len(peers),
		DownloadRate:  fmt.Sprintf("%f kBps", poolDownloadRate),
		UploadRate:    fmt.Sprintf("%f kBps", poolUploadRate),
		TotalDownload: fmt.Sprintf("%f kB", float64(poolDownloadTotal)/1000),
		TotalUpload:   fmt.Sprintf("%f kB", float64(poolUploadTotal)/1000),
	}

	resp := hal.NewResource(peerPoolStat, c.Request().RequestURI)
	resp.Embedded = e

	respB, _ := resp.MarshalJSON()
	return c.Blob(200, "application/hal+json", respB)

}

func (h *HTTPServe) peerID(c echo.Context) error {
	if h.PeerPool == nil {
		return c.String(http.StatusServiceUnavailable, "peerpool not available")
	}
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

	resp := hal.NewResource(peerDetail, c.Request().RequestURI)
	resp.AddNewLink("back", peersRoute.Path)
	respB, _ := resp.MarshalJSON()
	return c.Blob(200, "application/hal+json", respB)

}

func (h *HTTPServe) findPeerByHash(hashID string) (peer.Peer, error) {

	peers := h.PeerPool.Peers(peerpool.FilterConnected)
	for _, peerObj := range peers {
		if peerObj.GetID() == hashID {
			return peerObj, nil
		}
	}
	return nil, errors.New("not found")
}
