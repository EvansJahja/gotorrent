// Code generated by MockGen. DO NOT EDIT.
// Source: example.com/gotorrent/lib/core/adapter/peer (interfaces: Peer)

// Package mock_peer is a generated GoMock package.
package mock_peer

import (
	reflect "reflect"

	peer "example.com/gotorrent/lib/core/adapter/peer"
	domain "example.com/gotorrent/lib/core/domain"
	gomock "github.com/golang/mock/gomock"
)

// MockPeer is a mock of Peer interface.
type MockPeer struct {
	ctrl     *gomock.Controller
	recorder *MockPeerMockRecorder
}

// MockPeerMockRecorder is the mock recorder for MockPeer.
type MockPeerMockRecorder struct {
	mock *MockPeer
}

// NewMockPeer creates a new mock instance.
func NewMockPeer(ctrl *gomock.Controller) *MockPeer {
	mock := &MockPeer{ctrl: ctrl}
	mock.recorder = &MockPeerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPeer) EXPECT() *MockPeerMockRecorder {
	return m.recorder
}

// Choke mocks base method.
func (m *MockPeer) Choke() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Choke")
}

// Choke indicates an expected call of Choke.
func (mr *MockPeerMockRecorder) Choke() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Choke", reflect.TypeOf((*MockPeer)(nil).Choke))
}

// Connect mocks base method.
func (m *MockPeer) Connect() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Connect")
	ret0, _ := ret[0].(error)
	return ret0
}

// Connect indicates an expected call of Connect.
func (mr *MockPeerMockRecorder) Connect() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Connect", reflect.TypeOf((*MockPeer)(nil).Connect))
}

// DisconnectOnChokedChanged mocks base method.
func (m *MockPeer) DisconnectOnChokedChanged(arg0 func(bool)) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DisconnectOnChokedChanged", arg0)
}

// DisconnectOnChokedChanged indicates an expected call of DisconnectOnChokedChanged.
func (mr *MockPeerMockRecorder) DisconnectOnChokedChanged(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DisconnectOnChokedChanged", reflect.TypeOf((*MockPeer)(nil).DisconnectOnChokedChanged), arg0)
}

// DisconnectOnPiecesUpdatedChanged mocks base method.
func (m *MockPeer) DisconnectOnPiecesUpdatedChanged(arg0 func()) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DisconnectOnPiecesUpdatedChanged", arg0)
}

// DisconnectOnPiecesUpdatedChanged indicates an expected call of DisconnectOnPiecesUpdatedChanged.
func (mr *MockPeerMockRecorder) DisconnectOnPiecesUpdatedChanged(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DisconnectOnPiecesUpdatedChanged", reflect.TypeOf((*MockPeer)(nil).DisconnectOnPiecesUpdatedChanged), arg0)
}

// GetDownloadBytes mocks base method.
func (m *MockPeer) GetDownloadBytes() uint32 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDownloadBytes")
	ret0, _ := ret[0].(uint32)
	return ret0
}

// GetDownloadBytes indicates an expected call of GetDownloadBytes.
func (mr *MockPeerMockRecorder) GetDownloadBytes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDownloadBytes", reflect.TypeOf((*MockPeer)(nil).GetDownloadBytes))
}

// GetDownloadRate mocks base method.
func (m *MockPeer) GetDownloadRate() float32 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDownloadRate")
	ret0, _ := ret[0].(float32)
	return ret0
}

// GetDownloadRate indicates an expected call of GetDownloadRate.
func (mr *MockPeerMockRecorder) GetDownloadRate() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDownloadRate", reflect.TypeOf((*MockPeer)(nil).GetDownloadRate))
}

// GetMetadata mocks base method.
func (m *MockPeer) GetMetadata() (domain.Metadata, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetadata")
	ret0, _ := ret[0].(domain.Metadata)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMetadata indicates an expected call of GetMetadata.
func (mr *MockPeerMockRecorder) GetMetadata() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetadata", reflect.TypeOf((*MockPeer)(nil).GetMetadata))
}

// GetPeerID mocks base method.
func (m *MockPeer) GetPeerID() []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPeerID")
	ret0, _ := ret[0].([]byte)
	return ret0
}

// GetPeerID indicates an expected call of GetPeerID.
func (mr *MockPeerMockRecorder) GetPeerID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPeerID", reflect.TypeOf((*MockPeer)(nil).GetPeerID))
}

// GetState mocks base method.
func (m *MockPeer) GetState() peer.State {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetState")
	ret0, _ := ret[0].(peer.State)
	return ret0
}

// GetState indicates an expected call of GetState.
func (mr *MockPeerMockRecorder) GetState() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetState", reflect.TypeOf((*MockPeer)(nil).GetState))
}

// GetUploadBytes mocks base method.
func (m *MockPeer) GetUploadBytes() uint32 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUploadBytes")
	ret0, _ := ret[0].(uint32)
	return ret0
}

// GetUploadBytes indicates an expected call of GetUploadBytes.
func (mr *MockPeerMockRecorder) GetUploadBytes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUploadBytes", reflect.TypeOf((*MockPeer)(nil).GetUploadBytes))
}

// GetUploadRate mocks base method.
func (m *MockPeer) GetUploadRate() float32 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUploadRate")
	ret0, _ := ret[0].(float32)
	return ret0
}

// GetUploadRate indicates an expected call of GetUploadRate.
func (mr *MockPeerMockRecorder) GetUploadRate() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUploadRate", reflect.TypeOf((*MockPeer)(nil).GetUploadRate))
}

// Hostname mocks base method.
func (m *MockPeer) Hostname() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Hostname")
	ret0, _ := ret[0].(string)
	return ret0
}

// Hostname indicates an expected call of Hostname.
func (mr *MockPeerMockRecorder) Hostname() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Hostname", reflect.TypeOf((*MockPeer)(nil).Hostname))
}

// Interested mocks base method.
func (m *MockPeer) Interested() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Interested")
}

// Interested indicates an expected call of Interested.
func (mr *MockPeerMockRecorder) Interested() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Interested", reflect.TypeOf((*MockPeer)(nil).Interested))
}

// OnChokedChanged mocks base method.
func (m *MockPeer) OnChokedChanged(arg0 func(bool)) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "OnChokedChanged", arg0)
}

// OnChokedChanged indicates an expected call of OnChokedChanged.
func (mr *MockPeerMockRecorder) OnChokedChanged(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnChokedChanged", reflect.TypeOf((*MockPeer)(nil).OnChokedChanged), arg0)
}

// OnPiecesUpdatedChanged mocks base method.
func (m *MockPeer) OnPiecesUpdatedChanged(arg0 func()) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "OnPiecesUpdatedChanged", arg0)
}

// OnPiecesUpdatedChanged indicates an expected call of OnPiecesUpdatedChanged.
func (mr *MockPeerMockRecorder) OnPiecesUpdatedChanged(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnPiecesUpdatedChanged", reflect.TypeOf((*MockPeer)(nil).OnPiecesUpdatedChanged), arg0)
}

// OurPieces mocks base method.
func (m *MockPeer) OurPieces() domain.PieceList {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OurPieces")
	ret0, _ := ret[0].(domain.PieceList)
	return ret0
}

// OurPieces indicates an expected call of OurPieces.
func (mr *MockPeerMockRecorder) OurPieces() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OurPieces", reflect.TypeOf((*MockPeer)(nil).OurPieces))
}

// PieceRequests mocks base method.
func (m *MockPeer) PieceRequests() <-chan peer.PieceRequest {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PieceRequests")
	ret0, _ := ret[0].(<-chan peer.PieceRequest)
	return ret0
}

// PieceRequests indicates an expected call of PieceRequests.
func (mr *MockPeerMockRecorder) PieceRequests() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PieceRequests", reflect.TypeOf((*MockPeer)(nil).PieceRequests))
}

// RequestPiece mocks base method.
func (m *MockPeer) RequestPiece(arg0, arg1, arg2 uint32) <-chan []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RequestPiece", arg0, arg1, arg2)
	ret0, _ := ret[0].(<-chan []byte)
	return ret0
}

// RequestPiece indicates an expected call of RequestPiece.
func (mr *MockPeerMockRecorder) RequestPiece(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequestPiece", reflect.TypeOf((*MockPeer)(nil).RequestPiece), arg0, arg1, arg2)
}

// SetOurPiece mocks base method.
func (m *MockPeer) SetOurPiece(arg0 uint32) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetOurPiece", arg0)
}

// SetOurPiece indicates an expected call of SetOurPiece.
func (mr *MockPeerMockRecorder) SetOurPiece(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetOurPiece", reflect.TypeOf((*MockPeer)(nil).SetOurPiece), arg0)
}

// TellPieceCompleted mocks base method.
func (m *MockPeer) TellPieceCompleted(arg0 uint32) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "TellPieceCompleted", arg0)
}

// TellPieceCompleted indicates an expected call of TellPieceCompleted.
func (mr *MockPeerMockRecorder) TellPieceCompleted(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TellPieceCompleted", reflect.TypeOf((*MockPeer)(nil).TellPieceCompleted), arg0)
}

// TheirPieces mocks base method.
func (m *MockPeer) TheirPieces() domain.PieceList {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TheirPieces")
	ret0, _ := ret[0].(domain.PieceList)
	return ret0
}

// TheirPieces indicates an expected call of TheirPieces.
func (mr *MockPeerMockRecorder) TheirPieces() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TheirPieces", reflect.TypeOf((*MockPeer)(nil).TheirPieces))
}

// Unchoke mocks base method.
func (m *MockPeer) Unchoke() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Unchoke")
}

// Unchoke indicates an expected call of Unchoke.
func (mr *MockPeerMockRecorder) Unchoke() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unchoke", reflect.TypeOf((*MockPeer)(nil).Unchoke))
}

// Uninterested mocks base method.
func (m *MockPeer) Uninterested() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Uninterested")
}

// Uninterested indicates an expected call of Uninterested.
func (mr *MockPeerMockRecorder) Uninterested() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Uninterested", reflect.TypeOf((*MockPeer)(nil).Uninterested))
}
