package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	name   string
	Artist string
}

var (
	spotifyClientId     = "486808e1bcaf4c99a04d678ec687d805"
	spotifyClientSecret = "af514f94c2ba4908a245dbec9e17608e"
)

func main() {
	songs := getSongsFromTriplejAPI()
	err := updateSpotifyPlaylist(songs)
	if err != nil {
		fmt.Println("Could not update spotify playlist")
	}
}

func updateSpotifyPlaylist(songs []Song) error {

}

func getSongsFromTriplejAPI() []Song {

	var (
		triplejResponse TriplejResponse
		songs           []Song
	)

	url := "https://music.abcradio.net.au/api/v1/plays/search.json?station=triplej&limit=100&order=desc"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = json.Unmarshal(body, &triplejResponse)

	if err != nil {
		fmt.Println(err)
		return
	}

	for i := range triplejResponse.Items {
		recording := triplejResponse.Items[i].Recording
		songs = append(
			songs,
			Song{name: recording.Title, Artist: recording.Artists[0].Name},
		)
	}
	return songs
}
