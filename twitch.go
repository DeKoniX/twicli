package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
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

func (tw *TW) connect(url, oauth string) (body []byte, err error) {
	req, _ := http.NewRequest("GET", "https://api.twitch.tv/kraken/"+url, nil)
	req.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
	req.Header.Set("Client-ID", tw.ClientID)
	if oauth != "" {
		req.Header.Set("Authorization", "OAuth "+oauth)
	}
	resp, err := tw.HTTPClient.Do(req)
	if err != nil {
		return body, err
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}

	return body, nil
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

func (tw *TW) GetOnline(oauth string, page int) (streams []Stream, err error) {
	u := url.Values{}
	u.Set("limit", "10")
	u.Set("stream_type", "live")
	u.Set("offset", strconv.Itoa(page*10))
	body, err := tw.connect("streams/followed?"+u.Encode(), oauth)
	if err != nil {
		return streams, err
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

	return streams, nil
}

func (tw *TW) GetLive(lang string, page int) (streams []Stream, err error) {
	u := url.Values{}
	u.Set("limit", "10")
	u.Set("stream_type", "live")
	u.Set("offset", strconv.Itoa(page*10))
	var body []byte
	if lang == "" {
		body, err = tw.connect("streams?"+u.Encode(), "")
	} else {
		u.Set("language", lang)
		body, err = tw.connect("streams?"+u.Encode(), "")
	}
	if err != nil {
		return streams, err
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

	return streams, nil
}

func (tw *TW) GetSearch(search string, page int) (streams []Stream, err error) {
	u := url.Values{}
	u.Set("limit", "10")
	u.Set("query", search)
	u.Set("offset", strconv.Itoa(page*10))

	body, err := tw.connect("search/streams?"+u.Encode(), "")
	if err != nil {
		return streams, err
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

	return streams, nil
}

func getLength(timeStream time.Time) int {
	return int(time.Now().Unix() - timeStream.Unix())
}
