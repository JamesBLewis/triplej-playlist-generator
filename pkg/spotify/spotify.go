package spotify

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"

	"github.com/JamesBLewis/triplej-playlist-generator/pkg/telemetry"
)

const (
	ContentType = "application/json; charset=UTF-8"
	Market      = "AU"
)

type (
	Clienter interface {
		GetCurrentPlaylist(ctx context.Context, playlistId string) ([]Track, error)
		GetTrackBySongNameAndArtist(ctx context.Context, name string, artist []string) (Track, error)
		RemoveSongsFromPlaylist(ctx context.Context, songs []Track, playlistId string) error
		AddSongsToPlaylist(ctx context.Context, songs []string, playlistId string) error
	}

	Client struct {
		musicAPI     string
		accountAPI   string
		accessToken  string
		clientId     string
		clientSecret string
		refreshToken string
		httpClient   *http.Client
	}

	PlaylistTracks struct {
		Items []PlaylistTrackItem `json:"items"`
	}

	PlaylistTrackItem struct {
		Track Track `json:"track"`
	}

	Track struct {
		Uri string `json:"uri"`
	}

	TokenRefreshResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		ExpiresIn   int    `json:"expires_in"`
	}

	SearchTracksResponse struct {
		Tracks SearchTracks `json:"tracks"`
	}

	SearchTracks struct {
		Items []SearchTrackItem `json:"items"`
	}

	SearchTrackItem struct {
		Uri  string `json:"uri"`
		Name string `json:"name"`
	}
)

//go:generate mockgen -destination=mocks/spotify.go -source=spotify.go

func NewSpotifyClient(clientId, clientSecret, refreshToken string) Clienter {
	return &Client{
		musicAPI:     "https://api.spotify.com/v1",
		accountAPI:   "https://accounts.spotify.com/api",
		clientId:     clientId,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		accessToken:  "",
		// yes this is an arbitrary timeout I've pulled out of thin air
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Do wraps httpClient.Do and injects an access token into the request's header
func (sc *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// if no access token is set, refresh it.
	if sc.accessToken == "" {
		if err := sc.refreshAccessToken(ctx); err != nil {
			return nil, err
		}
	}

	// Set the Authorization header to use the new access token
	req.Header.Set("Authorization", "Bearer "+sc.accessToken)

	// Now send the request using the http.Client
	return sc.httpClient.Do(req)
}

func (sc *Client) refreshAccessToken(ctx context.Context) error {
	// Add a child span
	_, childSpan := otel.Tracer(telemetry.TracerName).Start(ctx, "refreshAccessToken")
	defer childSpan.End()
	fmt.Println("Refreshing Spotify access token...")

	encodedIdAndSecret := base64.StdEncoding.EncodeToString([]byte(sc.clientId + ":" + sc.clientSecret))
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", sc.refreshToken)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, sc.accountAPI+"/token", strings.NewReader(data.Encode()))
	if err != nil {
		return errors.Wrap(err, "failed to create new request")
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+encodedIdAndSecret)
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	res, err := sc.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to execute request")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code: %d", res.StatusCode)
	}

	tokenRefreshResponse := &TokenRefreshResponse{}
	if err := json.NewDecoder(res.Body).Decode(tokenRefreshResponse); err != nil {
		return errors.Wrap(err, "failed to unmarshal response body")
	}

	sc.accessToken = tokenRefreshResponse.AccessToken
	return nil
}

func (sc *Client) GetCurrentPlaylist(ctx context.Context, playlistId string) ([]Track, error) {
	// Add a child span
	ctx, childSpan := otel.Tracer(telemetry.TracerName).Start(ctx, "GetCurrentPlaylist")
	defer childSpan.End()
	requestUrl, err := url.JoinPath(sc.musicAPI, "playlists", playlistId, "tracks")
	if err != nil {
		return nil, errors.Wrap(err, "failed to construct request url")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestUrl, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new request")
	}

	// Add the fields and limit parameter to the request
	query := req.URL.Query()
	query.Add("fields", "items(track.uri)")
	query.Add("limit", "50")
	req.URL.RawQuery = query.Encode()

	res, err := sc.Do(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code: %d", res.StatusCode)
	}

	playlistTracks := &PlaylistTracks{}
	if err := json.NewDecoder(res.Body).Decode(playlistTracks); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response body")
	}

	var songs []Track
	for _, item := range playlistTracks.Items {
		songs = append(songs, Track{Uri: item.Track.Uri})
	}
	return songs, nil
}

func (sc *Client) GetTrackBySongNameAndArtist(ctx context.Context, name string, artists []string) (Track, error) {
	// Add a child span
	ctx, childSpan := otel.Tracer(telemetry.TracerName).Start(ctx, "GetTrackBySongNameAndArtist")
	defer childSpan.End()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sc.musicAPI+"/search", nil)
	if err != nil {
		return Track{}, errors.Wrap(err, "failed to create new request")
	}

	query := req.URL.Query()
	query.Add("q", fmt.Sprintf("%s %s", name, strings.Join(artists, ", ")))
	query.Add("type", "track")
	query.Add("market", Market)
	query.Add("limit", "1")
	req.URL.RawQuery = query.Encode()

	res, err := sc.Do(ctx, req)
	if err != nil {
		return Track{}, errors.Wrap(err, "failed to execute request")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Track{}, fmt.Errorf("invalid status code: %d", res.StatusCode)
	}

	searchTracksResponse := &SearchTracksResponse{}
	if err := json.NewDecoder(res.Body).Decode(searchTracksResponse); err != nil {
		return Track{}, errors.Wrap(err, "failed to unmarshal response body")
	}

	if len(searchTracksResponse.Tracks.Items) == 0 {
		return Track{}, fmt.Errorf("could not find track: %s %s", name, strings.Join(artists, ", "))
	}
	return Track{
		Uri: searchTracksResponse.Tracks.Items[0].Uri,
	}, nil
}

func (sc *Client) RemoveSongsFromPlaylist(ctx context.Context, songs []Track, playlistId string) error {
	// Add a child span
	ctx, childSpan := otel.Tracer(telemetry.TracerName).Start(ctx, "RemoveSongsFromPlaylist")
	defer childSpan.End()
	if len(songs) == 0 {
		return nil
	}

	// Create a struct to hold the tracks data
	type playlistData struct {
		Tracks []Track `json:"tracks"`
	}

	// Fill the struct with our songs data
	data := playlistData{Tracks: songs}

	// Marshal the struct into JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "failed to marshal songs")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/playlists/%s/tracks", sc.musicAPI, playlistId), bytes.NewBuffer(jsonData))
	if err != nil {
		return errors.Wrap(err, "failed to create new request")
	}

	req.Header.Set("Content-Type", ContentType)

	res, err := sc.Do(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to execute request")
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code: %d", res.StatusCode)
	}

	return nil
}

func (sc *Client) AddSongsToPlaylist(ctx context.Context, songs []string, playlistId string) error {
	// Add a child span
	ctx, childSpan := otel.Tracer(telemetry.TracerName).Start(ctx, "AddSongsToPlaylist")
	defer childSpan.End()
	if len(songs) == 0 {
		return nil
	}

	jsonList, err := json.Marshal(songs)
	if err != nil {
		return errors.Wrap(err, "failed to marshal songs")
	}
	jsonData := fmt.Sprintf(`{"uris":%s}`, string(jsonList))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/playlists/%s/tracks", sc.musicAPI, playlistId), bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		return errors.Wrap(err, "failed to create new request")
	}

	req.Header.Set("Content-Type", ContentType)

	res, err := sc.Do(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to execute request")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code: %d", res.StatusCode)
	}

	return nil
}
