package torrentdir

import (
	"io/fs"
	"path/filepath"

	"example.com/gotorrent/lib/core/adapter/localrepo"
	"example.com/gotorrent/lib/core/domain"
)

type FileRepo struct {
	BasePath string
}

var _ localrepo.LocalRepo = FileRepo{}

func (f FileRepo) GetFiles() []domain.File {
	var files []domain.File
	filepath.WalkDir(f.BasePath, func(path string, d fs.DirEntry, err error) error {
		// TODO unix only
		if d.Name()[0] == '.' {
			return nil
		}
		files = append(files, domain.File{Path: path})
		return nil
	})
	return files

}
