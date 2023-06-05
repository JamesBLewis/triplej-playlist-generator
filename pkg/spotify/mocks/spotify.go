// Code generated by MockGen. DO NOT EDIT.
// Source: spotify.go

// Package mock_spotify is a generated GoMock package.
package mock_spotify

import (
	context "context"
	reflect "reflect"

	spotify "github.com/JamesBLewis/triplej-playlist-generator/pkg/spotify"
	gomock "github.com/golang/mock/gomock"
)

// MockClienter is a mock of Clienter interface.
type MockClienter struct {
	ctrl     *gomock.Controller
	recorder *MockClienterMockRecorder
}

// MockClienterMockRecorder is the mock recorder for MockClienter.
type MockClienterMockRecorder struct {
	mock *MockClienter
}

// NewMockClienter creates a new mock instance.
func NewMockClienter(ctrl *gomock.Controller) *MockClienter {
	mock := &MockClienter{ctrl: ctrl}
	mock.recorder = &MockClienterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClienter) EXPECT() *MockClienterMockRecorder {
	return m.recorder
}

// AddSongsToPlaylist mocks base method.
func (m *MockClienter) AddSongsToPlaylist(ctx context.Context, songs []string, playlistId string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddSongsToPlaylist", ctx, songs, playlistId)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddSongsToPlaylist indicates an expected call of AddSongsToPlaylist.
func (mr *MockClienterMockRecorder) AddSongsToPlaylist(ctx, songs, playlistId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddSongsToPlaylist", reflect.TypeOf((*MockClienter)(nil).AddSongsToPlaylist), ctx, songs, playlistId)
}

// GetCurrentPlaylist mocks base method.
func (m *MockClienter) GetCurrentPlaylist(ctx context.Context, playlistId string) ([]spotify.Track, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCurrentPlaylist", ctx, playlistId)
	ret0, _ := ret[0].([]spotify.Track)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCurrentPlaylist indicates an expected call of GetCurrentPlaylist.
func (mr *MockClienterMockRecorder) GetCurrentPlaylist(ctx, playlistId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCurrentPlaylist", reflect.TypeOf((*MockClienter)(nil).GetCurrentPlaylist), ctx, playlistId)
}

// GetTrackBySongNameAndArtist mocks base method.
func (m *MockClienter) GetTrackBySongNameAndArtist(ctx context.Context, name string, artist []string) (spotify.Track, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTrackBySongNameAndArtist", ctx, name, artist)
	ret0, _ := ret[0].(spotify.Track)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTrackBySongNameAndArtist indicates an expected call of GetTrackBySongNameAndArtist.
func (mr *MockClienterMockRecorder) GetTrackBySongNameAndArtist(ctx, name, artist interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTrackBySongNameAndArtist", reflect.TypeOf((*MockClienter)(nil).GetTrackBySongNameAndArtist), ctx, name, artist)
}

// RemoveSongsFromPlaylist mocks base method.
func (m *MockClienter) RemoveSongsFromPlaylist(ctx context.Context, songs []spotify.Track, playlistId string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveSongsFromPlaylist", ctx, songs, playlistId)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveSongsFromPlaylist indicates an expected call of RemoveSongsFromPlaylist.
func (mr *MockClienterMockRecorder) RemoveSongsFromPlaylist(ctx, songs, playlistId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveSongsFromPlaylist", reflect.TypeOf((*MockClienter)(nil).RemoveSongsFromPlaylist), ctx, songs, playlistId)
}