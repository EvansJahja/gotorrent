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

			dir := path.Dir(pathToFile)
			os.MkdirAll(dir, os.ModePerm)
			f, _ := os.Create(pathToFile)
			f.Truncate(int64(p.Length))
			f.Close()
			wg.Done()

		}(p)
	}
	wg.Wait()

	fmt.Print("done")

}

func (f Files) GetLocalPiece(pieceNo uint32) []byte {
	pieceLength := f.Torrent.PieceLength

	skipBytes := int(pieceNo) * pieceLength

	numOfPieces := len(f.Torrent.Pieces) / 20
	if int(pieceNo) >= numOfPieces {
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
		//fmt.Printf("Read %s [%x]\n", pathToFile, skipBytes)
		if err != nil {
			panic(err)
		}
		skipBytes = 0
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

func (f Files) WritePieceToLocal(pieceNo int, pieceReader io.Reader, readOffset int64) (int, error) {

	pieceLength := f.Torrent.PieceLength

	skipBytes := pieceNo * pieceLength
	skipBytesDueToReader := 0
	skipBytesDueToReader += int(readOffset)
	skipBytes += skipBytesDueToReader

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
		fd.Close()
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

func (f Files) GetPiecesFromFile(relPath string) []uint32 {
	var bytesOffset []uint32
	lastByte := uint32(0)
	for _, fp := range f.Torrent.Files {
		bytesOffset = append(bytesOffset, lastByte)
		lastByte += uint32(fp.Length)
	}
	for i, fp := range f.Torrent.Files {
		s := f.GetRelativePath(fp.Path)
		if s != relPath {
			continue
		}
		// this is the file
		startByte := bytesOffset[i]
		lastByte := startByte + uint32(fp.Length) - 1 // if we are at the end of file list

		if i != len(f.Torrent.Files)-1 {
			// this is NOT the last file
			lastByte = bytesOffset[i+1]
		}

		return f.GetPiecesNoFromBytePos(startByte, lastByte)

	}

	return nil

}

func min(a uint32, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}
func max(a uint32, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

// Check whether test is inside [a,b]
func onSegment(a uint32, b uint32, test1 uint32, test2 uint32) bool {
	X := test1 <= max(a, b) && test2 >= min(a, b)
	Y := test2 <= max(a, b) && test2 >= min(a, b)
	return X || Y

}

func (f Files) GetPiecesNoFromBytePos(startPos uint32, endPos uint32) []uint32 {
	var pieces []uint32
	for i := 0; i < f.Torrent.PiecesCount(); i++ {
		startOfPiece := uint32(i * f.Torrent.PieceLength)
		endOfPiece := startOfPiece + uint32(f.Torrent.PieceLength) - 1

		if onSegment(startOfPiece, endOfPiece, startPos, endPos) {
			pieces = append(pieces, uint32(i))
		}
	}
	return pieces

}

func (f Files) VerifyLocalPiece(pieceNo uint32) bool {
	b := f.GetLocalPiece(pieceNo)

	hasher := sha1.New()
	hasher.Write(b)
	sumresult := hasher.Sum(nil)

	expResult := f.Torrent.Pieces[pieceNo*20 : pieceNo*20+20]
	if !bytes.Equal([]byte(expResult), sumresult) {
		fmt.Printf("corrupt %d\n", pieceNo)

		fmt.Printf("exp: %x\n", f.Torrent.Pieces[pieceNo*20:pieceNo*20+20])
		fmt.Printf("actual: %x\n", sumresult)
		return false

	}

	fmt.Printf("#%d OK\n", pieceNo)
	return true
}

func (f Files) CheckPieces(ourPieces domain.PieceList) (okPieces domain.PieceList, hasChanges bool) {
	okPieces = ourPieces

	for pieceNo := uint32(0); pieceNo < uint32(f.Torrent.PiecesCount()); pieceNo++ {
		if ourPieces.ContainPiece(pieceNo) {
			if ok := f.VerifyLocalPiece(pieceNo); !ok {
				okPieces.ResetPiece(pieceNo)
				hasChanges = true
			}
		}
	}
	return

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

func (f Files) GetRelativePath(p []string) string {
	return path.Join(p...)

}
