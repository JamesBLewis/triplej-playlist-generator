package spotify

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_AddSongsToPlaylist(t *testing.T) {

	testCtx := context.Background()

	type args struct {
		ctx        context.Context
		songs      []string
		playlistId string
	}

	t.Run("add song", func(t *testing.T) {

		args := args{
			testCtx,
			[]string{"spotify:track:2I66eI2j2ZfOe9q8TMLPbj"},
			"someplaylistid",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != fmt.Sprintf("/playlists/%s/tracks", args.playlistId) {
				t.Errorf("Expected to request '/playlists/%s/tracks', got: %s", args.playlistId, r.URL.Path)
			}

			w.WriteHeader(http.StatusCreated)
		}))
		defer server.Close()

		sc := &Client{
			accountAPI:   server.URL,
			musicAPI:     server.URL,
			accessToken:  "123",
			clientId:     "456",
			clientSecret: "321",
			refreshToken: "890",
			httpClient:   http.DefaultClient,
		}

		err := sc.AddSongsToPlaylist(args.ctx, args.songs, args.playlistId)

		require.NoError(t, err, "did not expect an error adding song to playlist")
	})
}

func TestClient_Do(t *testing.T) {
	t.Run("token already refreshed", func(t *testing.T) {

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/test" {
				t.Errorf("Expected to request '/test', got: %s", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		testRequest, _ := http.NewRequest("GET", fmt.Sprintf("%s/test", server.URL), nil)

		sc := &Client{
			accountAPI:   server.URL,
			musicAPI:     server.URL,
			accessToken:  "123",
			clientId:     "456",
			clientSecret: "321",
			refreshToken: "890",
			httpClient:   http.DefaultClient,
		}

		_, err := sc.Do(testRequest)
		require.NoError(t, err, "did not expect an error")
	})
}

func TestClient_GetCurrentPlaylist(t *testing.T) {

	testCtx := context.Background()

	type args struct {
		ctx        context.Context
		playlistId string
	}
	t.Run("fetch current playlist", func(t *testing.T) {

		args := args{
			testCtx,
			"somePlaylistId",
		}

		testTrack := "spotify:track:2I66eI2j2ZfOe9q8TMLPbj"

		want := []Track{{Uri: testTrack}}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != fmt.Sprintf("/playlists/%s/tracks", args.playlistId) {
				t.Errorf("Expected to request '/playlists/%s/tracks', got: %s", args.playlistId, r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"items":[{"track":{"uri":"%s"}}]}`, testTrack)))
		}))
		defer server.Close()

		sc := &Client{
			accountAPI:   server.URL,
			musicAPI:     server.URL,
			accessToken:  "someaccesstoken",
			clientId:     "123",
			clientSecret: "456",
			refreshToken: "789",
			httpClient:   http.DefaultClient,
		}

		got, err := sc.GetCurrentPlaylist(args.ctx, args.playlistId)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})
}

func TestClient_GetTrackBySongNameAndArtist(t *testing.T) {

	type args struct {
		ctx    context.Context
		name   string
		artist []string
	}

	testCtx := context.Background()

	t.Run("get track", func(t *testing.T) {

		args := args{
			testCtx,
			"The Duck Song",
			[]string{"The Duck"},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/search" {
				t.Errorf("Expected to request '/search', got: %s", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"tracks":{"items":[{"uri":"%s","name":"%s"}]}}`, "spotify:track:2I66eI2j2ZfOe9q8TMLPbj", "The Duck Song")))
		}))
		defer server.Close()

		sc := &Client{
			accountAPI:   server.URL,
			musicAPI:     server.URL,
			accessToken:  "someaccesstoken",
			clientId:     "123",
			clientSecret: "456",
			refreshToken: "789",
			httpClient:   http.DefaultClient,
		}

		want := Track{
			"spotify:track:2I66eI2j2ZfOe9q8TMLPbj",
		}

		got, err := sc.GetTrackBySongNameAndArtist(args.ctx, args.name, args.artist)
		require.NoError(t, err)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("GetTrackBySongNameAndArtist() got = %v, want %v", got, want)
		}
	})

}

func TestClient_RemoveSongsFromPlaylist(t *testing.T) {

	testCtx := context.Background()

	type args struct {
		ctx        context.Context
		songs      []Track
		playlistId string
	}

	t.Run("test delete", func(t *testing.T) {

		args := args{
			testCtx,
			[]Track{{"spotify:track:2I66eI2j2ZfOe9q8TMLPbj"}},
			"playlistId",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != fmt.Sprintf("/playlists/%s/tracks", args.playlistId) {
				t.Errorf("Expected to request: /playlists/%s/tracks\nGot: %s", args.playlistId, r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		sc := &Client{
			accountAPI:   server.URL,
			musicAPI:     server.URL,
			accessToken:  "someaccesstoken",
			clientId:     "123",
			clientSecret: "456",
			refreshToken: "789",
			httpClient:   http.DefaultClient,
		}

		err := sc.RemoveSongsFromPlaylist(args.ctx, args.songs, args.playlistId)
		require.NoError(t, err)
	})
}

func TestClient_refreshAccessToken(t *testing.T) {
	t.Run("fetch token", func(t *testing.T) {

		testAccessToken := "testAccessToken"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/token" {
				t.Errorf("Expected to request '/token', got: %s", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"access_token":"%s"}`, testAccessToken)))
		}))
		defer server.Close()

		sc := &Client{
			accountAPI:   server.URL,
			musicAPI:     server.URL,
			accessToken:  "",
			clientId:     "123",
			clientSecret: "456",
			refreshToken: "789",
			httpClient:   http.DefaultClient,
		}

		err := sc.refreshAccessToken()
		require.NoError(t, err, "no error expected when refreshing token")
		require.Equal(t, testAccessToken, sc.accessToken, "expected access token to be populated but it wasn't")
	})

	t.Run("unexpected status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/token" {
				t.Errorf("Expected to request '/token', got: %s", r.URL.Path)
			}

			w.WriteHeader(http.StatusBadGateway)
		}))
		defer server.Close()

		sc := &Client{
			accountAPI:   server.URL,
			musicAPI:     server.URL,
			accessToken:  "",
			clientId:     "123",
			clientSecret: "456",
			refreshToken: "789",
			httpClient:   http.DefaultClient,
		}

		err := sc.refreshAccessToken()
		require.Error(t, err, "error expected when refreshing token due to status code")
	})

	t.Run("invalid response body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/token" {
				t.Errorf("Expected to request '/token', got: %s", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`invalid response body`))
		}))
		defer server.Close()

		sc := &Client{
			accountAPI:   server.URL,
			musicAPI:     server.URL,
			accessToken:  "",
			clientId:     "123",
			clientSecret: "456",
			refreshToken: "789",
			httpClient:   http.DefaultClient,
		}

		err := sc.refreshAccessToken()
		require.Error(t, err, "error expected when refreshing token due to bad response body")
	})
}

func TestNewSpotifyClient(t *testing.T) {
	type args struct {
		clientId     string
		clientSecret string
		refreshToken string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "normal initialization",
			args: args{
				clientId:     "1234",
				clientSecret: "secret",
				refreshToken: "4321",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := &Client{
				musicAPI:     "https://api.spotify.com/v1",
				accountAPI:   "https://accounts.spotify.com/api",
				clientId:     tt.args.clientId,
				clientSecret: tt.args.clientSecret,
				refreshToken: tt.args.refreshToken,
				accessToken:  "",
				// yes this is an arbitrary timeout I've pulled out of thin air
				httpClient: &http.Client{Timeout: 10 * time.Second},
			}
			if got := NewSpotifyClient(tt.args.clientId, tt.args.clientSecret, tt.args.refreshToken); !reflect.DeepEqual(got, want) {
				t.Errorf("NewSpotifyClient() = %v, want %v", got, want)
			}
		})
	}
}
