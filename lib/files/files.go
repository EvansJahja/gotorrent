package files

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"example.com/gotorrent/lib/core/domain"
)

type Files struct {
	Torrent  domain.Torrent
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
			panic(err)
		}
		curPiece = append(curPiece, newBytes...)
		if len(curPiece) == pieceLength {
			break
		}

	}
	return curPiece

}

func (f Files) WriteSeeker(pieceNo int) io.WriteSeeker {
	return &writeSeekImpl{
		f:       f,
		pieceNo: pieceNo,
	}
}

type writeSeekImpl struct {
	pieceNo int
	f       Files
	cursor  int64
}

func (impl *writeSeekImpl) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	reader := bytes.NewReader(p)
	n, err = impl.f.WritePieceToLocal(impl.pieceNo, reader, impl.cursor)
	impl.cursor += int64(n)
	return
}

func (impl *writeSeekImpl) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		impl.cursor += offset
	case io.SeekStart:
		impl.cursor = offset
	case io.SeekEnd:
		impl.cursor = int64(impl.f.Torrent.PieceLength) + offset
	}
	return impl.cursor, nil
}

func (f Files) WritePieceToLocal(pieceNo int, pieceReader io.Reader, readOffset int64) (int, error) {

	pieceLength := f.Torrent.PieceLength

	skipBytes := pieceNo * pieceLength

	// If pieceReader is a seeker, it may want to give data with further progress
	/*
		seeker, ok := pieceReader.(io.ReadSeeker)
		if ok {
	*/

	skipBytesDueToReader := 0
	skipBytesDueToReader += int(readOffset)
	skipBytes += skipBytesDueToReader
	/*
		}
	*/

	numOfPieces := len(f.Torrent.Pieces) / 20
	if pieceNo >= numOfPieces {
		fmt.Printf("invalid piece no")
		return 0, errors.New("invalid piece no")
	}

	baseOfFile := 0
	cumRead := 0
	for fileIdx, fp := range f.Torrent.Files {
		if fp.Length <= skipBytes {
			skipBytes -= fp.Length
			baseOfFile += fp.Length
			continue
		}
		pathToFile := f.getAbsolutePath(fp.Path)
		fd, err := os.OpenFile(pathToFile, os.O_WRONLY, 0)
		if err != nil {
			panic(err)
		}
		fd.Truncate(int64(fp.Length))
		fd.Seek(int64(skipBytes), 1) // Absolute
		remainingToWrite := f.Torrent.Files[fileIdx].Length - skipBytes
		skipBytes = 0

		limitReader := io.LimitReader(pieceReader, int64(remainingToWrite))
		n, err := io.Copy(fd, limitReader)
		cumRead += int(n)
		if err != nil {
			return cumRead, err
		}
		if err == nil {
			if n == int64(remainingToWrite) {
				continue
			}
		}

		return cumRead, nil
	}
	return 0, nil
}

func (f Files) CheckFiles() {
	pieceNo := 2
	b := f.GetLocalPiece(pieceNo)

	hasher := sha1.New()
	hasher.Write(b)
	sumresult := hasher.Sum(nil)
	fmt.Printf("%x\n", f.Torrent.Pieces[pieceNo*20:pieceNo*20+20])
	fmt.Printf("%x\n", sumresult)

	os.Exit(1)
	/*
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
	*/

}

func (f Files) getAbsolutePath(p []string) string {
	var pathFragments []string
	pathFragments = append(pathFragments, f.BasePath)
	pathFragments = append(pathFragments, p...)

	pathToFile := path.Join(pathFragments...)
	return pathToFile
}
