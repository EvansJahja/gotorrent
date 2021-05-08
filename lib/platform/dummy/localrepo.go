package dummy

import (
	"example.com/gotorrent/lib/core/adapter/localrepo"
	"example.com/gotorrent/lib/core/domain"
)

type localrepoImpl struct{}

func LocalRepo() localrepo.LocalRepo {
	return localrepoImpl{}
}

var _ localrepo.LocalRepo = localrepoImpl{}

func (localrepoImpl) GetFiles() []domain.File {
	return []domain.File{
		{
			Path: "a",
		},
		{
			Path: "b",
		},
		{
			Path: "c",
		},
	}
}
