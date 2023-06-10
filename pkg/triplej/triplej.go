package triplej

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
)

const (
	abcRadioAPIBaseURL = "https://music.abcradio.net.au/api/v1/plays/search.json?station=triplej"
)

type Client struct{}

type RadioSong struct {
	Id      string
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
	Id      string   `json:"arid"`
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
		return nil, errors.Wrap(err, "creating request to ABC Radio musicAPI failed")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "GET request to ABC Radio musicAPI failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Received non-200 response code")
	}

	if err := json.NewDecoder(resp.Body).Decode(&triplejResponse); err != nil {
		return nil, errors.Wrap(err, "Decoding JSON response failed")
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

		songs = append(songs, RadioSong{Id: rec.Id, Name: rec.Title, Artists: artists})
	}
	return songs, nil
}
