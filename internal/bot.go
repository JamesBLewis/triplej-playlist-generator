package internal

import (
	"context"

	"github.com/pkg/errors"

	"github.com/JamesBLewis/triplej-playlist-generator/internal/config"
	"github.com/JamesBLewis/triplej-playlist-generator/pkg/log"
	"github.com/JamesBLewis/triplej-playlist-generator/pkg/spotify"
	"github.com/JamesBLewis/triplej-playlist-generator/pkg/triplej"
)

type Bot struct {
	spotifyClient     spotify.Clienter
	triplejClient     triplej.Clienter
	playlistSize      int
	spotifyPlaylistId string
	log               log.Log
}

func NewBot(config config.Config, logger log.Log) *Bot {
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

	recentTriplejSongs, err := b.triplejClient.FetchSongsFromTriplejAPI(ctx, b.playlistSize)
	if err != nil {
		return errors.Wrap(err, "Error fetching songs from TripleJ")
	}

	b.log.InfoContext(ctx, "Retrieved songs from triplej", "recentTriplejSongs", len(recentTriplejSongs))
	if len(recentTriplejSongs) == 0 {
		return errors.New("recentTriplejSongs contained 0 songs")
	}

	currentPlaylistSongs, err := b.spotifyClient.GetCurrentPlaylist(ctx, b.spotifyPlaylistId)
	if err != nil {
		return errors.Wrap(err, "Error fetching current spotify playlist")
	}
	b.log.InfoContext(ctx, "tracks found in the current spotify playlist", "currentPlaylistSongs", len(currentPlaylistSongs))

	lastPlayedSong, err := b.getTrackBySongNameAndArtist(ctx, recentTriplejSongs[0])
	if err != nil {
		return errors.Wrap(err, "Could not find last triplej song on spotify")
	}

	// exit if the last played song on the radio was also the most recently added song to this playlist
	if len(currentPlaylistSongs) > 0 && lastPlayedSong.Uri == currentPlaylistSongs[len(currentPlaylistSongs)-1].Uri {
		b.log.InfoContext(ctx, "Playlist is already up to date with triplej")
		return nil
	}
	b.log.InfoContext(ctx, "🤖diff found between playlist and triplej. updating playlist...")

	mappedSongs = append(mappedSongs, lastPlayedSong.Uri)
	err = b.updateSpotifyPlaylist(ctx, recentTriplejSongs, currentPlaylistSongs, mappedSongs)
	if err != nil {
		return errors.Wrap(err, "Error updating spotify playlist")
	}

	return nil
}

func (b *Bot) getTrackBySongNameAndArtist(ctx context.Context, song triplej.RadioSong) (spotify.Track, error) {
	b.log.InfoContext(ctx, "looking up song", "song", song)
	track, err := b.spotifyClient.GetTrackBySongNameAndArtist(ctx, song.Name, song.Artists)
	if err != nil {
		return spotify.Track{}, errors.Wrap(err, "failed to get track")
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

	// Calculate the number of songs to remove
	numToRemove := len(songsToAdd) + len(SpotifySongs) - b.playlistSize

	// If we need to remove songs, slice the SpotifySongs slice to get the songs to remove
	if numToRemove > 0 {
		songsToRemove = append(songsToRemove, SpotifySongs[:numToRemove]...)
	}

	if len(songsToRemove) > 0 {
		b.log.InfoContext(ctx, "removing songs from playlist...", "songsToRemove", len(songsToRemove))
		err := b.spotifyClient.RemoveSongsFromPlaylist(ctx, songsToRemove, b.spotifyPlaylistId)
		if err != nil {
			return errors.Wrap(err, "Error removing songs from playlist")
		}
	}

	b.log.InfoContext(ctx, "adding songs to playlist...", "songsToAdd", len(songsToAdd))
	err := b.spotifyClient.AddSongsToPlaylist(ctx, songsToAdd, b.spotifyPlaylistId)
	if err != nil {
		return errors.Wrap(err, "Error adding songs to playlist")
	}

	return nil
}
