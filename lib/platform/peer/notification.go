package peer

func (impl *peerImpl) OnChokedChanged(fn func(isChoked bool)) {

	impl.onChokedChangedFns = append(impl.onChokedChangedFns, fn)
}

func (impl *peerImpl) OnPiecesUpdatedChanged(fn func()) {
	impl.onPiecesChangedFns = append(impl.onPiecesChangedFns, fn)
}
func (impl *peerImpl) DisconnectOnChokedChanged(func(isChoked bool)) {

}
func (impl *peerImpl) DisconnectOnPiecesUpdatedChanged(func()) {

}
