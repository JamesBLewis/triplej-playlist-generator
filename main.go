package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

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

type SpotifyPlaylistTracksResponse struct {
	Tracks PlaylistTracks `json:"tracks"`
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

type Uri struct {
	Uri string `json:"uri"`
}

var (
	spotifyClientId     = ""
	spotifyClientSecret = ""
	spotifyAccessToken  = ""
	spotifyRefreshToken = ""
	spotifyPlaylistId   = ""
	playlistSize        = 20
)

func main() {
	//refreshSpotifyAccessToken()
	recentTriplejSongs := getSongsFromTriplejAPI()
	currentPlaylistSongs, err := getCurrentSpotifyPlayList()
	// check if the last song played on triplej is in our playlist already
	lastPlayedSong, err := getSpotifyTackBySongNameAndArtist(recentTriplejSongs[0].Name, recentTriplejSongs[0].Artist)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(currentPlaylistSongs) > 0 && lastPlayedSong.spotifyUri == currentPlaylistSongs[len(currentPlaylistSongs)-1].spotifyUri {
		fmt.Println("Exiting... playlist is already up to date with triplej")
		return
	}

	fmt.Println("diff found between playlist and triplej. updating playlist...")

	err = updateSpotifyPlaylist(recentTriplejSongs, currentPlaylistSongs)
	if err != nil {
		fmt.Println("Could not update spotify playlist")
	}
}

func refreshSpotifyAccessToken() {
	fmt.Println("refreshing spotify access token...")
	var (
		accessTokenRefreshURL = "https://accounts.spotify.com/api/token"
		method                = "POST"
		encodedIdAndSecret    = b64.StdEncoding.EncodeToString([]byte(spotifyClientId + ":" + spotifyClientSecret))
		client                = &http.Client{}
		data                  = url.Values{}
		tokenRefreshResponse  TokenRefreshResponse
	)
	fmt.Println(encodedIdAndSecret)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", spotifyRefreshToken)
	encodedData := data.Encode()

	req, err := http.NewRequest(method, accessTokenRefreshURL, strings.NewReader(encodedData))
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+encodedIdAndSecret)
	req.Header.Add("Content-Length", strconv.Itoa(len(encodedData)))

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = json.Unmarshal(body, &tokenRefreshResponse)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Fprintln(os.Stdout, "new access token: %s", tokenRefreshResponse.AccessToken)
}

func updateSpotifyPlaylist(triplejSongs []Song, SpotifySongs []Song) error {

	var (
		mappedSongs   []string
		songsToDelete []Track
	)

	for i := range triplejSongs[1:] {
		song, err := getSpotifyTackBySongNameAndArtist(triplejSongs[i].Name, triplejSongs[i].Artist)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		triplejSongs[i].spotifyUri = song.spotifyUri

		if len(SpotifySongs)-1-i < 0 {
			for j := range triplejSongs[i:] {
				tempSong, err := getSpotifyTackBySongNameAndArtist(triplejSongs[i+j].Name, triplejSongs[i+j].Artist)
				triplejSongs[i+j].spotifyUri = tempSong.spotifyUri
				if err != nil {
					fmt.Println(err)
					return nil
				}
			}
			break
		}

		if triplejSongs[i].spotifyUri == SpotifySongs[len(SpotifySongs)-1-i].spotifyUri {
			triplejSongs = triplejSongs[:i]
			break
		}
	}

	for i := range triplejSongs {
		if len(triplejSongs[len(triplejSongs)-1-i].spotifyUri) > 0 {
			mappedSongs = append(mappedSongs, triplejSongs[len(triplejSongs)-1-i].spotifyUri)
		}
	}
	err := addSongsToSpotifyPlaylist(mappedSongs)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for i := range SpotifySongs[:min(len(mappedSongs), len(SpotifySongs))] {
		if len(SpotifySongs[i].spotifyUri) > 0 {
			songsToDelete = append(songsToDelete, Track{Uri: SpotifySongs[i].spotifyUri})
		}
	}

	if len(mappedSongs)+len(SpotifySongs)-len(songsToDelete) > playlistSize {
		songsToRemove := len(mappedSongs) + len(SpotifySongs) - len(songsToDelete) - playlistSize
		numberCurrentlyBeingRemoved := len(songsToDelete)
		for i := 0; i < songsToRemove; i++ {
			songsToDelete = append(songsToDelete, Track{Uri: SpotifySongs[i+numberCurrentlyBeingRemoved].spotifyUri})
		}
	}

	err = removeSongsFromSpotifyPlaylist(songsToDelete)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return nil
}

func min(firstNumber int, secondNumber int) int {
	if firstNumber < secondNumber {
		return firstNumber
	} else {
		return secondNumber
	}
}

func removeSongsFromSpotifyPlaylist(songs []Track) error {
	var (
		requestUrl = fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", spotifyPlaylistId)
		method     = "DELETE"
		client     = &http.Client{}
		data       = url.Values{}
		jsonList   []byte
	)

	fmt.Printf("removing %d songs from spotify playlist...", len(songs))

	if len(songs) == 0 {
		return nil
	}

	jsonList, err := json.Marshal(songs)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", spotifyRefreshToken)
	jsondata := fmt.Sprintf(`{"tracks":%s}`, string(jsonList))

	req, err := http.NewRequest(method, requestUrl, bytes.NewBuffer([]byte(jsondata)))
	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+spotifyAccessToken)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	stringBody := string(body)
	fmt.Println(stringBody)
	return nil
}

func addSongsToSpotifyPlaylist(songs []string) error {
	fmt.Println("adding new songs to spotify playlist...")
	var (
		requestUrl = fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", spotifyPlaylistId)
		method     = "POST"
		client     = &http.Client{}
		data       = url.Values{}
		jsonList   []byte
	)

	jsonList, err := json.Marshal(songs)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", spotifyRefreshToken)
	jsondata := fmt.Sprintf(`{"uris":%s}`, string(jsonList))

	req, err := http.NewRequest(method, requestUrl, bytes.NewBuffer([]byte(jsondata)))
	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+spotifyAccessToken)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	stringBody := string(body)
	fmt.Println(stringBody)
	return nil
}

func getSongsFromTriplejAPI() []Song {
	fmt.Println("fetching recently played music from the triplej API...")

	var (
		triplejResponse TriplejResponse
		songs           []Song
		url             = fmt.Sprintf("https://music.abcradio.net.au/api/v1/plays/search.json?station=triplej&limit=%d&order=desc", playlistSize)
		method          = "GET"

		client = &http.Client{}
	)

	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return nil
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
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

func getCurrentSpotifyPlayList() ([]Song, error) {
	var (
		spotifySearchResponse PlaylistTracks
		url                   = fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", spotifyPlaylistId)
		method                = "GET"
		client                = &http.Client{}
		songs                 []Song
	)

	req, err := http.NewRequest(method, url, nil)
	req.Header.Set("Authorization", "Bearer "+spotifyAccessToken)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

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

func getSpotifyTackBySongNameAndArtist(name string, artist string) (Song, error) {
	fmt.Fprintln(os.Stdout, "looking up ", name, " by ", artist)

	var (
		spotifySearchResponse SpotifySearchTracksResponse
		url                   = "https://api.spotify.com/v1/search"
		method                = "GET"
		client                = &http.Client{}
	)

	req, err := http.NewRequest(method, url, nil)
	req.Header.Set("Authorization", "Bearer "+spotifyAccessToken)

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
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
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
