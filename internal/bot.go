package internal

import (
	"context"
	"log"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/JamesBLewis/triplej-playlist-generator/internal/config"
	"github.com/JamesBLewis/triplej-playlist-generator/pkg/spotify"
	"github.com/JamesBLewis/triplej-playlist-generator/pkg/triplej"
)

type Bot struct {
	spotifyClient     spotify.Clienter
	triplejClient     triplej.Clienter
	playlistSize      int
	spotifyPlaylistId string
	log               *zap.Logger
}

func NewBot(config config.Config, logger *zap.Logger) *Bot {
	spotifyClient := spotify.NewSpotifyClient(config.SpotifyClientId, config.SpotifyClientSecret, config.SpotifyRefreshToken)
	return &Bot{
		spotifyClient:     spotifyClient,
		triplejClient:     triplej.NewTiplejClient(),
		playlistSize:      config.PlaylistSize,
		spotifyPlaylistId: config.SpotifyPlaylistId,
		log:               logger,
	}
}

func (b *Bot) Run(ctx context.Context) error {
	var mappedSongs []string

	recentTriplejSongs, err := b.triplejClient.FetchSongsFromTriplejAPI(b.playlistSize)
	if err != nil {
		return errors.Wrap(err, "Error fetching songs from TripleJ")
	}
	b.log.Info("Retrieved songs from triplej", zap.Int("recentTriplejSongs", len(recentTriplejSongs)))
	if len(recentTriplejSongs) == 0 {
		return errors.New("recentTriplejSongs contained 0 songs")
	}

	currentPlaylistSongs, err := b.spotifyClient.GetCurrentPlaylist(ctx, b.spotifyPlaylistId)
	if err != nil {
		return errors.Wrap(err, "Error fetching current spotify playlist")
	}
	b.log.Info("tracks found in the current spotify playlist", zap.Int("currentPlaylistSongs", len(currentPlaylistSongs)))

	lastPlayedSong, err := b.getTrackBySongNameAndArtist(ctx, recentTriplejSongs[0])
	if err != nil {
		return errors.Wrap(err, "Could not find last triplej song on spotify")
	}

	// exit if the last played song on the radio was also the most recently added song to this playlist
	if len(currentPlaylistSongs) > 0 && lastPlayedSong.Uri == currentPlaylistSongs[len(currentPlaylistSongs)-1].Uri {
		b.log.Info("Playlist is already up to date with triplej")
		return nil
	}
	b.log.Info("🤖diff found between playlist and triplej. updating playlist...")

	mappedSongs = append(mappedSongs, lastPlayedSong.Uri)
	err = b.updateSpotifyPlaylist(ctx, recentTriplejSongs, currentPlaylistSongs, mappedSongs)
	if err != nil {
		return errors.Wrap(err, "Error updating spotify playlist")
	}

	return nil
}

func (b *Bot) getTrackBySongNameAndArtist(ctx context.Context, song triplej.RadioSong) (spotify.Track, error) {
	b.log.Info("looking up song", zap.Any("song", song))
	track, err := b.spotifyClient.GetTrackBySongNameAndArtist(ctx, song.Name, song.Artists)
	if err != nil {

		log.Printf("Error getting track: %v", err)
		return spotify.Track{}, err
	}
	return track, nil
}

func (b *Bot) updateSpotifyPlaylist(ctx context.Context, triplejSongs []triplej.RadioSong, SpotifySongs []spotify.Track, songsToAdd []string) error {
	var songsToRemove []spotify.Track

	for index, song := range triplejSongs[1:] {
		// skip if duplicate songs were returned by the triplej API
		// Note: in this context as we have sliced the first item off the list,
		// the index is actually 1 less than you expect.
		if song.Id == triplejSongs[index].Id {
			continue
		}

		tempSong, err := b.getTrackBySongNameAndArtist(ctx, song)
		if err != nil {
			continue
		}

		if len(SpotifySongs) > 0 && SpotifySongs[len(SpotifySongs)-1].Uri == tempSong.Uri {
			break
		}

		if len(tempSong.Uri) > 0 {
			// prepend item to slice
			songsToAdd = append([]string{tempSong.Uri}, songsToAdd...)
		}
	}

	b.log.Info("adding songs to playlist...", zap.Int("songsToAdd", len(songsToAdd)))
	err := b.spotifyClient.AddSongsToPlaylist(ctx, songsToAdd, b.spotifyPlaylistId)
	if err != nil {
		return errors.Wrap(err, "Error adding songs to playlist")
	}

	if len(songsToAdd)+len(SpotifySongs) > b.playlistSize {
		for i := 0; i < len(songsToAdd)+len(SpotifySongs)-b.playlistSize; i++ {
			if len(SpotifySongs[i].Uri) > 0 {
				songsToRemove = append(songsToRemove, spotify.Track{Uri: SpotifySongs[i].Uri})
			}
		}
	}

	if len(songsToRemove) > 0 {
		b.log.Info("removing songs from playlist...", zap.Int("songsToRemove", len(songsToRemove)))
		err = b.spotifyClient.RemoveSongsFromPlaylist(ctx, songsToRemove, b.spotifyPlaylistId)
		if err != nil {
			return errors.Wrap(err, "Error removing songs from playlist")
		}
	}

	return nil
}
