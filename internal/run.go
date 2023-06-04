package internal

import (
	"context"
	"errors"
	"log"

	"github.com/JamesBLewis/triplej-playlist-generator/cmd/config"
	"github.com/JamesBLewis/triplej-playlist-generator/pkg/spotify"
	"github.com/JamesBLewis/triplej-playlist-generator/pkg/triplej"
)

type Bot struct {
	spotifyClient     spotify.Clienter
	triplejClient     triplej.Clienter
	playlistSize      int
	spotifyPlaylistId string
}

func NewBot(config config.Config) *Bot {
	spotifyClient := spotify.NewSpotifyClient(config.SpotifyClientId, config.SpotifyClientSecret, config.SpotifyRefreshToken)
	return &Bot{
		spotifyClient:     spotifyClient,
		triplejClient:     triplej.NewTiplejClient(),
		playlistSize:      config.PlaylistSize,
		spotifyPlaylistId: config.SpotifyPlaylistId,
	}
}

func (b *Bot) Run(ctx context.Context) error {
	var mappedSongs []string

	recentTriplejSongs, err := b.triplejClient.FetchSongsFromTriplejAPI(b.playlistSize)
	if err != nil {
		log.Printf("Error fetching songs from TripleJ: %v", err)
		return err
	}
	if len(recentTriplejSongs) == 0 {
		err = errors.New("recentTriplejSongs contained 0 songs")
		return err
	}

	currentPlaylistSongs, err := b.spotifyClient.GetCurrentPlaylist(ctx, b.spotifyPlaylistId)
	if err != nil {
		log.Printf("Error fetching current spotify playlist: %v", err)
		return err
	}

	lastPlayedSong, err := b.getTrackBySongNameAndArtist(ctx, recentTriplejSongs[0])
	if err != nil {
		log.Printf("Could not find last triplej song on spotify: %v", err)
		return err
	}

	// exit if the last played song on the radio was also the most recently added song to this playlist
	if len(currentPlaylistSongs) > 0 && lastPlayedSong.Uri == currentPlaylistSongs[len(currentPlaylistSongs)-1].Uri {
		log.Println("Exiting... playlist is already up to date with triplej")
		return nil
	}

	log.Println("🤖diff found between playlist and triplej. updating playlist...")

	mappedSongs = append(mappedSongs, lastPlayedSong.Uri)
	err = b.updateSpotifyPlaylist(ctx, recentTriplejSongs, currentPlaylistSongs, mappedSongs)
	if err != nil {
		log.Printf("Error updating spotify playlist: %v", err)
		return err
	}

	return nil
}

func (b *Bot) getTrackBySongNameAndArtist(ctx context.Context, song triplej.RadioSong) (spotify.Track, error) {
	log.Printf("looking up: %s", song.Name)
	track, err := b.spotifyClient.GetTrackBySongNameAndArtist(ctx, song.Name, song.Artists)
	if err != nil {

		log.Printf("Error getting track: %v", err)
		return spotify.Track{}, err
	}
	return track, nil
}

func (b *Bot) updateSpotifyPlaylist(ctx context.Context, triplejSongs []triplej.RadioSong, SpotifySongs []spotify.Track, mappedSongs []string) error {
	var songsToDelete []spotify.Track

	for _, song := range triplejSongs[1:] {
		tempSong, err := b.getTrackBySongNameAndArtist(ctx, song)
		if err != nil {
			continue
		}

		if len(SpotifySongs) > 0 && SpotifySongs[len(SpotifySongs)-1].Uri == tempSong.Uri {
			break
		}

		if len(tempSong.Uri) > 0 {
			// prepend item to slice
			mappedSongs = append([]string{tempSong.Uri}, mappedSongs...)
		}
	}

	log.Printf("adding %d songs to playlist...", len(mappedSongs))
	err := b.spotifyClient.AddSongsToPlaylist(ctx, mappedSongs, b.spotifyPlaylistId)
	if err != nil {
		log.Printf("Error adding songs to playlist: %v", err)
		return err
	}

	if len(mappedSongs)+len(SpotifySongs) > b.playlistSize {
		for i := 0; i < len(mappedSongs)+len(SpotifySongs)-b.playlistSize; i++ {
			if len(SpotifySongs[i].Uri) > 0 {
				songsToDelete = append(songsToDelete, spotify.Track{Uri: SpotifySongs[i].Uri})
			}
		}
	}

	if len(songsToDelete) > 0 {
		log.Printf("removing %d songs from playlist...", len(songsToDelete))
		err = b.spotifyClient.RemoveSongsFromPlaylist(ctx, songsToDelete, b.spotifyPlaylistId)
		if err != nil {
			log.Printf("Error removing songs from playlist: %v", err)
			return err
		}
	}

	return nil
}
