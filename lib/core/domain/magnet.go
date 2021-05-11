package domain

import (
	"encoding/hex"
	"net/url"
	"strings"
)

type Magnet struct{ Url *url.URL }

func (m *Magnet) Trackers() []*url.URL {
	trackers := m.Url.Query()["tr"]
	var resp []*url.URL
	for _, t := range trackers {
		trackerU, err := url.Parse(t)
		if err != nil {
			continue
		}
		resp = append(resp, trackerU)
	}
	return resp
}
func (m *Magnet) InfoHash() []byte {
	infoHashStr := strings.Split(m.Url.Query()["xt"][0], ":")[2]
	infoHash, _ := hex.DecodeString(infoHashStr)
	return infoHash

}
