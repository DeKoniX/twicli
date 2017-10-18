package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"regexp"

	"strings"

	"github.com/gizak/termui"
)

func (app *Application) updateStreamList(update bool, search string) {
	var err error
	var strs []string

	if update {
		app.UI.parNotiHelp.Text = "[Обновляю список стримов](fg-red)"
		termui.Render(app.UI.parNotiHelp)
		time.Sleep(10 * time.Millisecond)
		app.UI.parNotiHelp.Text = helpText
		app.Streams, err = app.getStreams(app.StreamType, app.DB, search, app.StreamPage)
		if err != nil {
			fmt.Println("Ошибка получения списка стримов: ", err)
			os.Exit(2)
		}

		app.StreamID = 0
		if search != "" {
			app.UI.parNotiHelp.Text = helpText + "\n Поиск: [" + app.Search + "](fg-blue)"
		}
	}
	for id, stream := range app.Streams {
		if id == app.StreamID {
			strs = append(strs, "["+stream.DisplayName+"](fg-white,bg-green)")
		} else {
			strs = append(strs, stream.DisplayName)
		}
	}
	app.UI.lsStreams.Items = strs
	if len(app.Streams) > 0 {
		app.UI.parName.Text = app.Streams[app.StreamID].Status
		app.UI.parGame.Text = app.Streams[app.StreamID].Game
		app.UI.parViewers.Text = strconv.Itoa(app.Streams[app.StreamID].Viewers)
		app.UI.parLength.Text = videoLen(app.Streams[app.StreamID].Length)
	}
	termui.Render(termui.Body)
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}

func videoLen(len int) (strLength string) {
	var hour, min, second int
	if len > 60 {
		min = len / 60
		second = len % 60

		if min > 59 {
			hour = min / 60
			min = min % 60

			strLength = fmt.Sprintf("Часов: %d, Минуты: %d, ", hour, min)
		} else {
			strLength = fmt.Sprintf("Минуты: %d, ", min)
		}
		strLength = strLength + fmt.Sprintf("Секунды: %d", second)
	} else {
		strLength = fmt.Sprintf("Секунды: %d", len)
	}

	return strLength
}

func (app *Application) getStreams(streamsType int, dataBase DB, search string, page int) (streams []Stream, err error) {
	switch streamsType {
	case 1:
		streams, err = app.TW.GetLive("", page)
		if err != nil {
			return streams, err
		}
	case 2:
		streams, err = app.TW.GetLive("ru", page)
		if err != nil {
			return streams, err
		}
	case 3:
		streams, err = app.TW.GetSearch(search, page)
		if err != nil {
			return streams, err
		}
	case 0:
		accessTokenRow, err := dataBase.SelectAccessToken()
		if err != nil {
			return streams, err
		}
		if accessTokenRow.accessToken == "" {
			u, _ := url.Parse("https://api.twitch.tv/kraken/oauth2/authorize")
			q := u.Query()
			q.Set("client_id", clientID)
			q.Set("redirect_uri", redirectURI)
			q.Set("response_type", "token")
			q.Set("scope", "user_read")
			u.RawQuery = q.Encode()
			fmt.Println("Open url: ", u.String())

			openBrowser(u.String())

			l, err := StartHttpServer()
			if err != nil {
				return streams, err
			}
			for ShutdownServer == false {
				time.Sleep(1 * time.Second)
			}
			l.Close()
			accessTokenRow, err = dataBase.SelectAccessToken()
			if err != nil {
				return streams, err
			}
		}
		streams, err = app.TW.GetOnline(accessTokenRow.accessToken, page)
		if err != nil {
			return streams, err
		}
	}
	return streams, nil
}

func (app *Application) runStreamlink(quality string) {
	if len(app.Streams) > 0 {
		if app.Cmd != nil {
			app.Cmd.Process.Kill()
			app.Cmd = nil
			app.UI.parStreamOn.Text = ""
			app.StreamNowName = ""
			termui.Render(app.UI.parStreamOn)
			time.Sleep(10 * time.Millisecond)
		}
		app.UI.parNotiHelp.Text = "[Запускаю streamlink](fg-red)"
		termui.Render(app.UI.parNotiHelp)
		app.StreamNowName = app.Streams[app.StreamID].DisplayName
		app.UI.parStreamOn.Text = "[" + app.StreamNowName + "](fg-red)"
		termui.Render(app.UI.parStreamOn)
		if quality == "" {
			app.Cmd = exec.Command("streamlink", "-p", "mpv --fs", "--default-stream", "720p,720p60,best,source", app.Streams[app.StreamID].URL)
		} else {
			app.Cmd = exec.Command("streamlink", "-p", "mpv --fs", app.Streams[app.StreamID].URL, quality)
		}
		var out bytes.Buffer
		app.Cmd.Stdout = &out
		err := app.Cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			app.Cmd.Wait()
			f, _ := os.Create("out")
			f.Write(out.Bytes())
			app.updateStreamList(false, app.Search)
		}()
	}
}

func (app *Application) getStreamlinkQuality() (quality []string, err error) {
	cmd := exec.Command("streamlink", app.Streams[app.StreamID].URL)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return quality, err
	}
	str := out.String()

	var qualityRegexp = regexp.MustCompile("Available streams: (.*)")
	var deleteWordRegexp = regexp.MustCompile(`\ ?\([a-zA-Z]{1,10},?\ ?[a-zA-Z]{1,10}\)`)
	r := qualityRegexp.FindStringSubmatch(str)
	if len(r) < 1 {
		return quality, errors.New("ERR Get quality stream")
	}
	str = deleteWordRegexp.ReplaceAllString(r[1], "")
	quality = strings.Split(str, ", ")

	return quality, nil
}
