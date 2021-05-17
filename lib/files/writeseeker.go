package files

import (
	"bytes"
	"io"
)

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
