package files

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"example.com/gotorrent/lib/core/domain"
)

type Files struct {
	Torrent  *domain.Torrent
	BasePath string
}

func (f Files) CreateFiles() {

	paths := f.Torrent.Files

	var wg sync.WaitGroup

	for _, p := range paths {
		wg.Add(1)
		go func(p domain.FileInfo) {

			pathToFile := f.getAbsolutePath(p.Path)

			//buf := make([]byte, p.Length)
			dir := path.Dir(pathToFile)
			os.MkdirAll(dir, os.ModePerm)
			f, _ := os.Create(pathToFile)
			f.Truncate(int64(p.Length))
			f.Close()
			//os.WriteFile(pathToFile, buf, os.ModePerm&0666)
			wg.Done()

		}(p)
	}
	wg.Wait()

	fmt.Print("done")

}

func (f Files) GetLocalPiece(pieceNo int) []byte {
	pieceLength := f.Torrent.PieceLength

	skipBytes := pieceNo * pieceLength

	numOfPieces := len(f.Torrent.Pieces) / 20
	if pieceNo >= numOfPieces {
		fmt.Printf("invalid piece no")
		return nil
	}

	curPiece := make([]byte, 0, pieceLength)

	for _, fp := range f.Torrent.Files {
		if fp.Length <= skipBytes {
			skipBytes -= fp.Length
			continue
		}
		pathToFile := f.getAbsolutePath(fp.Path)
		fd, err := os.Open(pathToFile)
		fd.Seek(int64(skipBytes), 1) // Absolute
		if err != nil {
			panic(err)
		}
		limReader := io.LimitReader(fd, int64(pieceLength-len(curPiece)))

		newBytes, err := io.ReadAll(limReader)
		if err != nil {
			return nil
		}
		curPiece = append(curPiece, newBytes...)
		if len(curPiece) == pieceLength {
			break
		}

	}
	return curPiece

}

func (f Files) WritePieceToLocal(pieceNo int, pieceReader io.Reader) {

	pieceLength := f.Torrent.PieceLength

	skipBytes := pieceNo * pieceLength

	numOfPieces := len(f.Torrent.Pieces) / 20
	if pieceNo >= numOfPieces {
		fmt.Printf("invalid piece no")
		return
	}

	//pieceReader := bytes.NewReader(piece)

	for _, fp := range f.Torrent.Files {
		if fp.Length <= skipBytes {
			skipBytes -= fp.Length
			continue
		}
		pathToFile := f.getAbsolutePath(fp.Path)
		fmt.Printf("Write to %s\n", pathToFile)
		fd, err := os.OpenFile(pathToFile, os.O_WRONLY, 0)
		fd.Seek(int64(skipBytes), 1) // Absolute
		skipBytes = 0
		if err != nil {
			panic(err)
		}
		remainingToWrite := fp.Length - skipBytes

		limitReader := io.LimitReader(pieceReader, int64(remainingToWrite))
		io.Copy(fd, limitReader)

	}
}

func (f Files) CheckFiles() {
	// Todo

	v := []byte(f.Torrent.Pieces)

	pieceLength := f.Torrent.PieceLength

	curPiece := make([]byte, 0, pieceLength)

	Z := func(a []byte) {
		//
		if len(a) != pieceLength {
			print("x")
		}

	}

	for _, p := range f.Torrent.Files {
		pathToFile := f.getAbsolutePath(p.Path)
		fd, err := os.Open(pathToFile)
		if err != nil {
			panic(err)
		}

		src := bufio.NewReader(fd)

		for {
			readBuf := io.LimitReader(src, int64(pieceLength-len(curPiece)))
			newBytes, err := ioutil.ReadAll(readBuf)
			if len(newBytes) == 0 {
				break
			}

			if err != nil {
				fmt.Print(err)
				return
			}
			//fmt.Println(i)

			curPiece = append(curPiece, newBytes...)
			//copy(curPiece[len(curPiece):], newBytes)

			if len(curPiece) >= pieceLength {
				//fmt.Print(len(curPiece))
				Z(curPiece)
				curPiece = make([]byte, 0, pieceLength)
				continue
			}
		}

	}
	Z(curPiece)

	fmt.Print(v)

}

func (f Files) getAbsolutePath(p []string) string {
	var pathFragments []string
	pathFragments = append(pathFragments, f.BasePath)
	pathFragments = append(pathFragments, p...)

	pathToFile := path.Join(pathFragments...)
	return pathToFile
}
