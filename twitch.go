package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type TW struct {
	ClientID    string
	RedirectURI string
	HTTPClient  *http.Client
}

func TWInit(clientID, redirectURI string) *TW {
	return &TW{
		ClientID:    clientID,
		HTTPClient:  &http.Client{},
		RedirectURI: redirectURI,
	}
}

func (tw *TW) connect(url, oauth string) (body []byte) {
	req, _ := http.NewRequest("GET", "https://api.twitch.tv/kraken/"+url, nil)
	req.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
	req.Header.Set("Client-ID", tw.ClientID)
	if oauth != "" {
		req.Header.Set("Authorization", "OAuth "+oauth)
	}
	resp, _ := tw.HTTPClient.Do(req)
	body, _ = ioutil.ReadAll(resp.Body)

	return body
}

type Stream struct {
	DisplayName string
	Game        string
	Name        string
	Status      string
	URL         string
	Viewers     int
	Date        time.Time
	Length      int
}

func (tw *TW) GetOnline(oauth string) (streams []Stream) {
	u := url.Values{}
	u.Set("limit", "10")
	u.Set("stream_type", "live")
	body := tw.connect("streams/followed?"+u.Encode(), oauth)

	type jsonTW struct {
		Streams []struct {
			Channel struct {
				DisplayName string `json:"display_name"`
				Name        string `json:"name"`
				URL         string `json:"url"`
				Status      string `json:"status"`
			}
			CreatedAt string `json:"created_at"`
			Game      string `json:"game"`
			Viewers   int    `json:"viewers"`
		}
	}

	var jsontw jsonTW
	json.Unmarshal(body, &jsontw)

	for _, stream := range jsontw.Streams {
		twTime, err := time.Parse(time.RFC3339, stream.CreatedAt)
		if err != nil {
			twTime = time.Now()
		}
		streams = append(streams, Stream{
			DisplayName: stream.Channel.DisplayName,
			Game:        stream.Game,
			Name:        stream.Channel.Name,
			Status:      stream.Channel.Status,
			URL:         stream.Channel.URL,
			Viewers:     stream.Viewers,
			Date:        twTime,
			Length:      getLength(twTime),
		})
	}

	return streams
}

func (tw *TW) GetLive(lang string) (streams []Stream) {
	u := url.Values{}
	u.Set("limit", "10")
	u.Set("stream_type", "live")
	var body []byte
	if lang == "" {
		body = tw.connect("streams?"+u.Encode(), "")
	} else {
		u.Set("language", lang)
		body = tw.connect("streams?"+u.Encode(), "")
	}

	type jsonTW struct {
		Streams []struct {
			Channel struct {
				DisplayName string `json:"display_name"`
				Name        string `json:"name"`
				URL         string `json:"url"`
				Status      string `json:"status"`
			}
			CreatedAt string `json:"created_at"`
			Game      string `json:"game"`
			Viewers   int    `json:"viewers"`
		}
	}

	var jsontw jsonTW
	json.Unmarshal(body, &jsontw)

	for _, stream := range jsontw.Streams {
		twTime, err := time.Parse(time.RFC3339, stream.CreatedAt)
		if err != nil {
			twTime = time.Now()
		}
		streams = append(streams, Stream{
			DisplayName: stream.Channel.DisplayName,
			Game:        stream.Game,
			Name:        stream.Channel.Name,
			Status:      stream.Channel.Status,
			URL:         stream.Channel.URL,
			Viewers:     stream.Viewers,
			Date:        twTime,
			Length:      getLength(twTime),
		})
	}

	return streams
}

func (tw *TW) GetSearch(search string) (streams []Stream) {
	u := url.Values{}
	u.Set("limit", "10")
	u.Set("query", search)

	body := tw.connect("search/streams?"+u.Encode(), "")

	type jsonTW struct {
		Streams []struct {
			Channel struct {
				DisplayName string `json:"display_name"`
				Name        string `json:"name"`
				URL         string `json:"url"`
				Status      string `json:"status"`
			}
			CreatedAt string `json:"created_at"`
			Game      string `json:"game"`
			Viewers   int    `json:"viewers"`
		}
	}

	var jsontw jsonTW
	json.Unmarshal(body, &jsontw)

	for _, stream := range jsontw.Streams {
		twTime, err := time.Parse(time.RFC3339, stream.CreatedAt)
		if err != nil {
			twTime = time.Now()
		}
		streams = append(streams, Stream{
			DisplayName: stream.Channel.DisplayName,
			Game:        stream.Game,
			Name:        stream.Channel.Name,
			Status:      stream.Channel.Status,
			URL:         stream.Channel.URL,
			Viewers:     stream.Viewers,
			Date:        twTime,
			Length:      getLength(twTime),
		})
	}

	return streams
}

func getLength(timeStream time.Time) int {
	return int(time.Now().Unix() - timeStream.Unix())
}
