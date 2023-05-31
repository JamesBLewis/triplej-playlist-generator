package config

import (
	"errors"
	"log"
	"os"
	"strconv"
)

type Config struct {
	SpotifyClientId     string
	SpotifyClientSecret string
	SpotifyAccessToken  string
	SpotifyRefreshToken string
	SpotifyPlaylistId   string
	PlaylistSize        int
}

func Load() (Config, error) {
	spotifyClientId := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyPlaylistId := os.Getenv("SPOTIFY_PLAYLIST_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	spotifyRefreshToken := os.Getenv("SPOTIFY_REFRESH_TOKEN")
	playlistSize, err := strconv.Atoi(os.Getenv("PLAYLIST_SIZE"))
	if err != nil {
		log.Println("PlaylistSize was invalid")
		return Config{}, err
	}

	config := Config{
		SpotifyPlaylistId:   spotifyPlaylistId,
		PlaylistSize:        playlistSize,
		SpotifyClientId:     spotifyClientId,
		SpotifyClientSecret: spotifyClientSecret,
		SpotifyRefreshToken: spotifyRefreshToken,
	}

	err = validateConfig(config)
	if err != nil {
		log.Println("one or more config fields were invalid")
		return Config{}, err
	}

	return config, nil
}

func validateConfig(config Config) error {
	if len(config.SpotifyPlaylistId) == 0 {
		return errors.New("empty SpotifyPlaylistId")
	}
	if config.PlaylistSize < 1 {
		return errors.New("playlist size was smaller then 1")
	}
	if len(config.SpotifyClientId) == 0 {
		return errors.New("empty SpotifyClientId")
	}
	if len(config.SpotifyClientSecret) == 0 {
		return errors.New("empty SpotifyClientSecret")
	}
	if len(config.SpotifyRefreshToken) == 0 {
		return errors.New("empty SpotifyRefreshToken")
	}
	return nil
}
