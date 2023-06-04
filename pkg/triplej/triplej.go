package triplej

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

const (
	abcRadioAPIBaseURL = "https://music.abcradio.net.au/api/v1/plays/search.json?station=triplej"
)

type Client struct{}

type RadioSong struct {
	Name    string
	Artists []string
}

type triplejResponse struct {
	Items []item `json:"items"`
}

type item struct {
	Recording recording `json:"recording"`
}

type recording struct {
	Title   string   `json:"title"`
	Artists []artist `json:"artists"`
}

type artist struct {
	Name string `json:"name"`
}

type Clienter interface {
	FetchSongsFromTriplejAPI(playlistSize int) ([]RadioSong, error)
}

//go:generate mockgen -destination=mocks/triplej.go -source=triplej.go

func NewTiplejClient() Client {
	return Client{}
}

func (Client) FetchSongsFromTriplejAPI(playlistSize int) ([]RadioSong, error) {
	log.Printf("Fetching recently played music from Triple J...")
	var (
		triplejResponse triplejResponse
		songs           []RadioSong
		abcUrl          = abcRadioAPIBaseURL + "&limit=" + strconv.Itoa(playlistSize) + "&order=desc"
	)

	if playlistSize < 0 {
		return []RadioSong(nil), errors.New("invalid playlist size")
	}
	songs = make([]RadioSong, 0, playlistSize)

	req, err := http.NewRequest("GET", abcUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request to ABC Radio musicAPI failed: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET request to ABC Radio musicAPI failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Received non-200 response code")
	}

	if err := json.NewDecoder(resp.Body).Decode(&triplejResponse); err != nil {
		return nil, fmt.Errorf("Decoding JSON response failed: %v", err)
	}

	for _, item := range triplejResponse.Items {
		var artists []string
		rec := item.Recording
		if len(rec.Artists) == 0 {
			log.Println("Warning: No artist information in the response")
			continue
		}

		for _, artist := range rec.Artists {
			artists = append(artists, artist.Name)
		}

		songs = append(songs, RadioSong{Name: rec.Title, Artists: artists})
	}

	log.Printf("Retrieved %d songs from Triple J", len(songs))
	return songs, nil
}
