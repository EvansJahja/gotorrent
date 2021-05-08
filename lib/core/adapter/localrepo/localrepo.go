package localrepo

import "example.com/gotorrent/lib/core/domain"

type LocalRepo interface {
	GetFiles() []domain.File
}
