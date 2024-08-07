package internal

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/JamesBLewis/triplej-playlist-generator/pkg/log"
	mock_spotify "github.com/JamesBLewis/triplej-playlist-generator/pkg/spotify/mocks"

	"github.com/JamesBLewis/triplej-playlist-generator/pkg/spotify"
	"github.com/JamesBLewis/triplej-playlist-generator/pkg/triplej"
	mock_triplej "github.com/JamesBLewis/triplej-playlist-generator/pkg/triplej/mocks"
)

func TestBot_Run(t *testing.T) {
	testCtx := context.Background()

	type args struct {
		ctx context.Context
	}

	t.Run("empty playlist", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSpotifyClient := mock_spotify.NewMockClienter(ctrl)
		mockTriplejClient := mock_triplej.NewMockClienter(ctrl)

		args := args{
			testCtx,
		}

		b := &Bot{
			spotifyClient:     mockSpotifyClient,
			triplejClient:     mockTriplejClient,
			playlistSize:      30,
			spotifyPlaylistId: "1234",
			log:               log.NewLogger(),
		}

		// mock logic
		triplejSongs := []triplej.RadioSong{
			{
				Id:   "2",
				Name: "latest song",
				Artists: []string{
					"artist 0",
					"artist 1",
				},
			},
			{
				Id:   "1",
				Name: "middle song",
				Artists: []string{
					"only artist",
				},
			},
			{
				Id:   "0",
				Name: "oldest song",
			},
		}
		mockTriplejClient.EXPECT().FetchSongsFromTriplejAPI(args.ctx, b.playlistSize).Return(triplejSongs, nil)
		mockSpotifyClient.EXPECT().GetCurrentPlaylist(args.ctx, b.spotifyPlaylistId).Return([]spotify.Track(nil), nil)
		mockSpotifyClient.EXPECT().GetTrackBySongNameAndArtist(args.ctx, triplejSongs[0].Name, triplejSongs[0].Artists).Return(spotify.Track{Uri: "uri:song0"}, nil)
		mockSpotifyClient.EXPECT().GetTrackBySongNameAndArtist(args.ctx, triplejSongs[1].Name, triplejSongs[1].Artists).Return(spotify.Track{Uri: "uri:song1"}, nil)
		mockSpotifyClient.EXPECT().GetTrackBySongNameAndArtist(args.ctx, triplejSongs[2].Name, triplejSongs[2].Artists).Return(spotify.Track{Uri: "uri:song2"}, nil)

		mockSpotifyClient.EXPECT().AddSongsToPlaylist(args.ctx, []string{"uri:song2", "uri:song1", "uri:song0"}, b.spotifyPlaylistId).Return(nil)

		err := b.Run(args.ctx)
		require.NoError(t, err)
	})

	t.Run("full playlist", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSpotifyClient := mock_spotify.NewMockClienter(ctrl)
		mockTriplejClient := mock_triplej.NewMockClienter(ctrl)

		args := args{
			testCtx,
		}

		b := &Bot{
			spotifyClient:     mockSpotifyClient,
			triplejClient:     mockTriplejClient,
			playlistSize:      3,
			spotifyPlaylistId: "1234",
			log:               log.NewLogger(),
		}

		currentTracks := []spotify.Track{{Uri: "uri:oldSong1"}, {Uri: "uri:oldSong2"}, {Uri: "uri:oldSong3"}}

		// mock logic
		triplejSongs := []triplej.RadioSong{
			{
				Id:   "2",
				Name: "latest song",
				Artists: []string{
					"artist 0",
					"artist 1",
				},
			},
			{
				Id:   "1",
				Name: "middle song",
				Artists: []string{
					"only artist",
				},
			},
			{
				Id:   "0",
				Name: "oldest song",
			},
		}
		mockTriplejClient.EXPECT().FetchSongsFromTriplejAPI(args.ctx, b.playlistSize).Return(triplejSongs, nil)
		mockSpotifyClient.EXPECT().GetCurrentPlaylist(args.ctx, b.spotifyPlaylistId).Return(currentTracks, nil)
		mockSpotifyClient.EXPECT().GetTrackBySongNameAndArtist(args.ctx, triplejSongs[0].Name, triplejSongs[0].Artists).Return(spotify.Track{Uri: "uri:song0"}, nil)
		mockSpotifyClient.EXPECT().GetTrackBySongNameAndArtist(args.ctx, triplejSongs[1].Name, triplejSongs[1].Artists).Return(spotify.Track{Uri: "uri:song1"}, nil)
		mockSpotifyClient.EXPECT().GetTrackBySongNameAndArtist(args.ctx, triplejSongs[2].Name, triplejSongs[2].Artists).Return(spotify.Track{Uri: "uri:song2"}, nil)

		mockSpotifyClient.EXPECT().AddSongsToPlaylist(args.ctx, []string{"uri:song2", "uri:song1", "uri:song0"}, b.spotifyPlaylistId).Return(nil)

		mockSpotifyClient.EXPECT().RemoveSongsFromPlaylist(args.ctx, currentTracks, b.spotifyPlaylistId)

		err := b.Run(args.ctx)
		require.NoError(t, err)
	})

	t.Run("duplicate triplej songs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSpotifyClient := mock_spotify.NewMockClienter(ctrl)
		mockTriplejClient := mock_triplej.NewMockClienter(ctrl)

		args := args{
			testCtx,
		}

		b := &Bot{
			spotifyClient:     mockSpotifyClient,
			triplejClient:     mockTriplejClient,
			playlistSize:      3,
			spotifyPlaylistId: "1234",
			log:               log.NewLogger(),
		}

		currentTracks := []spotify.Track{{Uri: "uri:oldSong1"}, {Uri: "uri:oldSong2"}, {Uri: "uri:oldSong3"}}

		// mock logic
		triplejSongs := []triplej.RadioSong{
			{
				Id:   "2",
				Name: "latest song",
				Artists: []string{
					"artist 0",
					"artist 1",
				},
			},
			{
				Id:   "1",
				Name: "duplicate song",
				Artists: []string{
					"only artist",
				},
			},
			{
				Id:   "1",
				Name: "duplicate song",
				Artists: []string{
					"only artist",
				},
			},
		}
		mockTriplejClient.EXPECT().FetchSongsFromTriplejAPI(args.ctx, b.playlistSize).Return(triplejSongs, nil)
		mockSpotifyClient.EXPECT().GetCurrentPlaylist(args.ctx, b.spotifyPlaylistId).Return(currentTracks, nil)
		mockSpotifyClient.EXPECT().GetTrackBySongNameAndArtist(args.ctx, triplejSongs[0].Name, triplejSongs[0].Artists).Return(spotify.Track{Uri: "uri:song0"}, nil)
		mockSpotifyClient.EXPECT().GetTrackBySongNameAndArtist(args.ctx, triplejSongs[1].Name, triplejSongs[1].Artists).Return(spotify.Track{Uri: "uri:song1"}, nil)

		mockSpotifyClient.EXPECT().AddSongsToPlaylist(args.ctx, []string{"uri:song1", "uri:song0"}, b.spotifyPlaylistId).Return(nil)

		mockSpotifyClient.EXPECT().RemoveSongsFromPlaylist(args.ctx, currentTracks[:2], b.spotifyPlaylistId)

		err := b.Run(args.ctx)
		require.NoError(t, err)
	})

	t.Run("existing playlist larger current playlistSize value", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSpotifyClient := mock_spotify.NewMockClienter(ctrl)
		mockTriplejClient := mock_triplej.NewMockClienter(ctrl)

		args := args{
			testCtx,
		}

		b := &Bot{
			spotifyClient:     mockSpotifyClient,
			triplejClient:     mockTriplejClient,
			playlistSize:      1,
			spotifyPlaylistId: "1234",
			log:               log.NewLogger(),
		}

		currentTracks := []spotify.Track{{Uri: "uri:oldSong1"}, {Uri: "uri:oldSong2"}, {Uri: "uri:oldSong3"}}

		// mock logic
		triplejSongs := []triplej.RadioSong{
			{
				Name: "latest song",
				Artists: []string{
					"artist 0",
					"artist 1",
				},
			},
		}
		mockTriplejClient.EXPECT().FetchSongsFromTriplejAPI(args.ctx, b.playlistSize).Return(triplejSongs, nil)
		mockSpotifyClient.EXPECT().GetCurrentPlaylist(args.ctx, b.spotifyPlaylistId).Return(currentTracks, nil)
		mockSpotifyClient.EXPECT().GetTrackBySongNameAndArtist(args.ctx, triplejSongs[0].Name, triplejSongs[0].Artists).Return(spotify.Track{Uri: "uri:latestsong"}, nil)
		mockSpotifyClient.EXPECT().AddSongsToPlaylist(args.ctx, []string{"uri:latestsong"}, b.spotifyPlaylistId).Return(nil)

		mockSpotifyClient.EXPECT().RemoveSongsFromPlaylist(args.ctx, currentTracks, b.spotifyPlaylistId)

		err := b.Run(args.ctx)
		require.NoError(t, err)
	})

	t.Run("existing playlist smaller than current playlistSize value", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSpotifyClient := mock_spotify.NewMockClienter(ctrl)
		mockTriplejClient := mock_triplej.NewMockClienter(ctrl)

		args := args{
			testCtx,
		}

		b := &Bot{
			spotifyClient:     mockSpotifyClient,
			triplejClient:     mockTriplejClient,
			playlistSize:      4,
			spotifyPlaylistId: "1234",
			log:               log.NewLogger(),
		}

		currentTracks := []spotify.Track{{Uri: "uri:oldSong1"}, {Uri: "uri:oldSong2"}}

		// mock logic
		triplejSongs := []triplej.RadioSong{
			{
				Id:   "2",
				Name: "latest song",
				Artists: []string{
					"artist 0",
					"artist 1",
				},
			},
			{
				Id:   "1",
				Name: "middle song",
				Artists: []string{
					"only artist",
				},
			},
			{
				Id:   "0",
				Name: "oldest song",
			},
		}
		mockTriplejClient.EXPECT().FetchSongsFromTriplejAPI(args.ctx, b.playlistSize).Return(triplejSongs, nil)
		mockSpotifyClient.EXPECT().GetCurrentPlaylist(args.ctx, b.spotifyPlaylistId).Return(currentTracks, nil)
		mockSpotifyClient.EXPECT().GetTrackBySongNameAndArtist(args.ctx, triplejSongs[0].Name, triplejSongs[0].Artists).Return(spotify.Track{Uri: "uri:latestsong"}, nil)
		mockSpotifyClient.EXPECT().GetTrackBySongNameAndArtist(args.ctx, triplejSongs[1].Name, triplejSongs[1].Artists).Return(spotify.Track{Uri: "uri:oldSong2"}, nil)

		mockSpotifyClient.EXPECT().AddSongsToPlaylist(args.ctx, []string{"uri:latestsong"}, b.spotifyPlaylistId).Return(nil)

		err := b.Run(args.ctx)
		require.NoError(t, err)
	})

	t.Run("up to date playlist", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSpotifyClient := mock_spotify.NewMockClienter(ctrl)
		mockTriplejClient := mock_triplej.NewMockClienter(ctrl)

		args := args{
			testCtx,
		}

		b := &Bot{
			spotifyClient:     mockSpotifyClient,
			triplejClient:     mockTriplejClient,
			playlistSize:      3,
			spotifyPlaylistId: "1234",
			log:               log.NewLogger(),
		}

		currentTracks := []spotify.Track{{Uri: "uri:oldSong1"}, {Uri: "uri:oldSong2"}, {Uri: "uri:song0"}}

		// mock logic
		triplejSongs := []triplej.RadioSong{
			{
				Name: "latest song",
				Artists: []string{
					"artist 0",
					"artist 1",
				},
			},
			{
				Name: "middle song",
				Artists: []string{
					"only artist",
				},
			},
			{
				Name: "oldest song",
			},
		}
		mockTriplejClient.EXPECT().FetchSongsFromTriplejAPI(args.ctx, b.playlistSize).Return(triplejSongs, nil)
		mockSpotifyClient.EXPECT().GetCurrentPlaylist(args.ctx, b.spotifyPlaylistId).Return(currentTracks, nil)
		mockSpotifyClient.EXPECT().GetTrackBySongNameAndArtist(args.ctx, triplejSongs[0].Name, triplejSongs[0].Artists).Return(spotify.Track{Uri: "uri:song0"}, nil)

		err := b.Run(args.ctx)
		require.NoError(t, err)
	})

	t.Run("empty response from triplej", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockSpotifyClient := mock_spotify.NewMockClienter(ctrl)
		mockTriplejClient := mock_triplej.NewMockClienter(ctrl)

		args := args{
			testCtx,
		}

		b := &Bot{
			spotifyClient:     mockSpotifyClient,
			triplejClient:     mockTriplejClient,
			playlistSize:      3,
			spotifyPlaylistId: "1234",
			log:               log.NewLogger(),
		}

		mockTriplejClient.EXPECT().FetchSongsFromTriplejAPI(args.ctx, b.playlistSize).Return([]triplej.RadioSong{}, nil)

		err := b.Run(args.ctx)
		require.Error(t, err)
	})
}
