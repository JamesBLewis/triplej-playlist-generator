package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	spotifyClientId     string
	spotifyClientSecret string
	spotifyAccessToken  string
	spotifyRefreshToken string
	spotifyPlaylistId   string
	playlistSize        int
}

type TriplejResponse struct {
	Items []Item `json:"items"`
}

type Item struct {
	Recording Recording `json:"recording"`
}

type Recording struct {
	Title   string   `json:"title"`
	Artists []Artist `json:"artists"`
}

type Artist struct {
	Name string `json:"name"`
}

type Song struct {
	Name       string
	Artist     string
	spotifyUri string
}

type TokenRefreshResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	ExpiresIn   int    `json:"expires_in"`
}

type PlaylistTracks struct {
	Items []SpotifyPlaylistTrackItem `json:"items"`
}

type SpotifyPlaylistTrackItem struct {
	Track Track `json:"track"`
}

type Track struct {
	Uri  string `json:"uri"`
	Name string `json:"name"`
}

type SpotifySearchTracksResponse struct {
	Tracks SearchTracks `json:"tracks"`
}

type SearchTracks struct {
	Items []SpotifySearchTrackItem `json:"items"`
}

type SpotifySearchTrackItem struct {
	Uri  string `json:"uri"`
	Name string `json:"name"`
}

func main() {
	config := configFromEnv()
	fmt.Println("🤖Triplej Bot is running...")

	config.spotifyAccessToken = refreshSpotifyAccessToken(config)
	recentTriplejSongs := getSongsFromTriplejAPI(config)
	currentPlaylistSongs, err := getCurrentSpotifyPlayList(config)
	// check if the last song played on triplej is in our playlist already
	lastPlayedSong, err := getSpotifyTackBySongNameAndArtist(recentTriplejSongs[0].Name, recentTriplejSongs[0].Artist, config)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(currentPlaylistSongs) > 0 && lastPlayedSong.spotifyUri == currentPlaylistSongs[len(currentPlaylistSongs)-1].spotifyUri {
		fmt.Println("Exiting... playlist is already up to date with triplej")
		return
	}

	fmt.Println("🤖diff found between playlist and triplej. updating playlist...")

	err = updateSpotifyPlaylist(recentTriplejSongs, currentPlaylistSongs, config)
	if err != nil {
		fmt.Println("Could not update spotify playlist")
	}
	fmt.Println("🤖Done.")
}

func configFromEnv() Config {
	spotifyClientId := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyPlaylistId := os.Getenv("SPOTIFY_PLAYLIST_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	spotifyRefreshToken := os.Getenv("SPOTIFY_REFRESH_TOKEN")
	playlistSize, err := strconv.Atoi(os.Getenv("PLAYLIST_SIZE"))
	if err != nil {
		log.Fatal("playlistSize was invalid")
	}

	config := Config{
		spotifyPlaylistId:   spotifyPlaylistId,
		playlistSize:        playlistSize,
		spotifyClientId:     spotifyClientId,
		spotifyClientSecret: spotifyClientSecret,
		spotifyRefreshToken: spotifyRefreshToken,
	}

	err = validateConfig(config)
	if err != nil {
		log.Fatal("one or more config fields were invalid", err)
	}

	return config
}

func validateConfig(config Config) error {
	if len(config.spotifyPlaylistId) == 0 {
		return errors.New("empty spotifyPlaylistId")
	}
	if config.playlistSize < 1 {
		return errors.New("playlist size was smaller then 1")
	}
	if len(config.spotifyClientId) == 0 {
		return errors.New("empty spotifyClientId")
	}
	if len(config.spotifyClientSecret) == 0 {
		return errors.New("empty spotifyClientSecret")
	}
	if len(config.spotifyRefreshToken) == 0 {
		return errors.New("empty spotifyRefreshToken")
	}
	return nil
}

func refreshSpotifyAccessToken(config Config) string {
	fmt.Println("refreshing spotify access token...")
	var (
		accessTokenRefreshURL = "https://accounts.spotify.com/api/token"
		method                = "POST"
		encodedIdAndSecret    = b64.StdEncoding.EncodeToString([]byte(config.spotifyClientId + ":" + config.spotifyClientSecret))
		client                = &http.Client{}
		data                  = url.Values{}
		tokenRefreshResponse  TokenRefreshResponse
	)

	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", config.spotifyRefreshToken)
	encodedData := data.Encode()

	req, err := http.NewRequest(method, accessTokenRefreshURL, strings.NewReader(encodedData))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+encodedIdAndSecret)
	req.Header.Add("Content-Length", strconv.Itoa(len(encodedData)))

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		log.Fatal("invalid status code", res.StatusCode)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Print(err)
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	err = json.Unmarshal(body, &tokenRefreshResponse)

	if err != nil {
		fmt.Println(err)
		return ""
	}

	return tokenRefreshResponse.AccessToken
}

func updateSpotifyPlaylist(triplejSongs []Song, SpotifySongs []Song, config Config) error {

	var (
		mappedSongs   []string
		songsToDelete []Track
	)

	for i := range triplejSongs[1:] {
		tempSong, err := getSpotifyTackBySongNameAndArtist(triplejSongs[i].Name, triplejSongs[i].Artist, config)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if len(SpotifySongs) > 0 && SpotifySongs[len(SpotifySongs)-1].spotifyUri == tempSong.spotifyUri {
			break
		}
		triplejSongs[i].spotifyUri = tempSong.spotifyUri
	}

	for i := range triplejSongs {
		if len(triplejSongs[len(triplejSongs)-1-i].spotifyUri) > 0 {
			mappedSongs = append(mappedSongs, triplejSongs[len(triplejSongs)-1-i].spotifyUri)
		}
	}
	err := addSongsToSpotifyPlaylist(mappedSongs, config)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if len(mappedSongs)+len(SpotifySongs) > config.playlistSize {
		for i := 0; i < len(mappedSongs)+len(SpotifySongs)-config.playlistSize; i++ {
			if len(SpotifySongs[i].spotifyUri) > 0 {
				songsToDelete = append(songsToDelete, Track{Uri: SpotifySongs[i].spotifyUri})
			}
		}
	}

	err = removeSongsFromSpotifyPlaylist(songsToDelete, config)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return nil
}

func removeSongsFromSpotifyPlaylist(songs []Track, config Config) error {
	var (
		requestUrl = fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", config.spotifyPlaylistId)
		method     = "DELETE"
		client     = &http.Client{}
		data       = url.Values{}
		jsonList   []byte
	)

	fmt.Printf("removing %d songs from spotify playlist...\n", len(songs))

	if len(songs) == 0 {
		return nil
	}

	jsonList, err := json.Marshal(songs)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", config.spotifyRefreshToken)
	jsonData := fmt.Sprintf(`{"tracks":%s}`, string(jsonList))

	req, err := http.NewRequest(method, requestUrl, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+config.spotifyAccessToken)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(res.Body)

	return nil
}

func addSongsToSpotifyPlaylist(songs []string, config Config) error {
	fmt.Printf("adding %d new songs to spotify playlist...\n", len(songs))
	var (
		requestUrl = fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", config.spotifyPlaylistId)
		method     = "POST"
		client     = &http.Client{}
		data       = url.Values{}
		jsonList   []byte
	)

	jsonList, err := json.Marshal(songs)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", config.spotifyRefreshToken)
	jsonData := fmt.Sprintf(`{"uris":%s}`, string(jsonList))

	req, err := http.NewRequest(method, requestUrl, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+config.spotifyAccessToken)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(res.Body)

	return nil
}

func getSongsFromTriplejAPI(config Config) []Song {
	fmt.Println("fetching recently played music from the triplej API...")

	var (
		triplejResponse TriplejResponse
		songs           []Song
		abcUrl          = fmt.Sprintf("https://music.abcradio.net.au/api/v1/plays/search.json?station=triplej&limit=%d&order=desc", config.playlistSize)
		method          = "GET"

		client = &http.Client{}
	)

	req, err := http.NewRequest(method, abcUrl, nil)

	if err != nil {
		fmt.Println(err)
		return nil
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if res.StatusCode != http.StatusOK {
		log.Fatal("invalid status code", res.StatusCode)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	err = json.Unmarshal(body, &triplejResponse)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	for i := range triplejResponse.Items {
		recording := triplejResponse.Items[i].Recording
		songs = append(
			songs,
			Song{Name: recording.Title, Artist: recording.Artists[0].Name},
		)
	}
	return songs
}

func getCurrentSpotifyPlayList(config Config) ([]Song, error) {
	var (
		spotifySearchResponse PlaylistTracks
		spotifyUrl            = fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", config.spotifyPlaylistId)
		method                = "GET"
		client                = &http.Client{}
		songs                 []Song
	)

	req, err := http.NewRequest(method, spotifyUrl, nil)
	req.Header.Set("Authorization", "Bearer "+config.spotifyAccessToken)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	err = json.Unmarshal(body, &spotifySearchResponse)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	for i := range spotifySearchResponse.Items {
		songUri := spotifySearchResponse.Items[i].Track.Uri
		songName := spotifySearchResponse.Items[i].Track.Name
		songs = append(
			songs,
			Song{spotifyUri: songUri, Name: songName},
		)
	}
	return songs, nil
}

func getSpotifyTackBySongNameAndArtist(name string, artist string, config Config) (Song, error) {
	fmt.Println("looking up:", name, "by", artist)

	var (
		spotifySearchResponse SpotifySearchTracksResponse
		spotifyUrl            = "https://api.spotify.com/v1/search"
		method                = "GET"
		client                = &http.Client{}
	)

	req, err := http.NewRequest(method, spotifyUrl, nil)
	req.Header.Set("Authorization", "Bearer "+config.spotifyAccessToken)

	query := req.URL.Query()
	query.Add("q", fmt.Sprintf("track:%s artist:%s", name, artist))
	query.Add("type", "track,artist")
	query.Add("market", "AU")
	query.Add("limit", "1")
	req.URL.RawQuery = query.Encode()

	if err != nil {
		fmt.Println(err)
		return Song{}, err
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return Song{}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return Song{}, err
	}

	err = json.Unmarshal(body, &spotifySearchResponse)

	if err != nil {
		fmt.Println(err)
		return Song{}, err
	}

	if len(spotifySearchResponse.Tracks.Items) == 0 {
		fmt.Println("could not find track")
		return Song{}, err
	}
	return Song{
		spotifyUri: spotifySearchResponse.Tracks.Items[0].Uri,
		Name:       spotifySearchResponse.Tracks.Items[0].Name,
	}, nil
}
