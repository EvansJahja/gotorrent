package main

import (
	"example.com/gotorrent/lib/domain"
	"example.com/gotorrent/lib/files"
	"example.com/gotorrent/lib/persistence"
	"example.com/gotorrent/lib/runner"
)

func main() {

	location := "/home/evans/torrent/test/"

	t := domain.Torrent{}

	p := persistence.New(location, &t)
	p.SetUpdateChan(t.GetUpdatedChan())

	f := files.Files{Torrent: &t}
	f.CreateFiles()

	r := runner.Runner{
		Torrent: &t,
		Files:   f,
	}

	//r.SetupTracker()
	r.Start()

	//f.GetLocalPiece(240)
	//f.WritePieceToLocal(1, []byte{0x0, 0x0, 0x0, 0x0})
	//f.WritePieceToLocal(1, []byte{0xDE, 0xAD, 0xBE, 0xEF})
	//x := f.GetLocalPiece(1)
	//fmt.Printf("%x", x[0:4])
	//f.CheckFiles()

	//t.Print()
	//return

	//t.Start()
	select {}

}
