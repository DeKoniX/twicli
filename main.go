package main

import (
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"runtime"

	"strconv"

	"time"

	"bytes"
	"os"

	"github.com/gizak/termui"
)

const clientID = ""
const redirectURI = "http://localhost:5454"

const helpText = "[<up>, <down>] - вверх, вниз по списку доступных стримов [r] - обновить список стримов [q] - выйти из приложения\n [f] - показать тех к кому ты подписан [o] - показать 10 самых топовых стримов"

var Streams []Stream
var StreamID int = 0
var StreamFol bool = true

func main() {
	var err error
	var cmd *exec.Cmd

	dataBase, err := initDB()
	if err != nil {
		log.Panic(err)
	}
	Streams = getStreams(StreamFol, dataBase)
	var strs []string
	for id, stream := range Streams {
		if id == 0 {
			strs = append(strs, "["+stream.DisplayName+"](fg-white,bg-green)")
		} else {
			strs = append(strs, stream.DisplayName)
		}
	}
	err = termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	lsStreams := termui.NewList()
	lsStreams.Items = strs
	lsStreams.BorderLabel = "Стримы:"
	lsStreams.Height = 12

	parName := termui.NewPar(Streams[0].Status)
	parName.Height = 3
	parName.BorderLabel = "Наименование:"

	parGame := termui.NewPar(Streams[0].Game)
	parGame.Height = 3
	parGame.BorderLabel = "Игра:"

	parViewers := termui.NewPar(strconv.Itoa(Streams[0].Viewers))
	parViewers.Height = 3
	parViewers.BorderLabel = "Смотрят:"

	parLength := termui.NewPar(videoLen(Streams[0].Length))
	parLength.Height = 3
	parLength.BorderLabel = "Идет уже:"

	parNotiHelp := termui.NewPar(helpText)
	parNotiHelp.Height = 3
	parNotiHelp.Border = false

	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(3, 0, lsStreams),
			termui.NewCol(9, 0, parName, parGame, parViewers, parLength),
		),
		termui.NewRow(
			termui.NewCol(12, 0, parNotiHelp),
		),
	)

	termui.Body.Align()

	termui.Render(termui.Body)
	termui.Handle("/sys/kbd/q", func(event termui.Event) {
		if cmd != nil {
			cmd.Process.Kill()
			cmd = nil
		} else {
			termui.StopLoop()
		}
	})
	// /sys/kbd/<down>
	// /sys/kbd/<up>
	termui.Handle("/sys/kbd/<down>", func(event termui.Event) {
		strs = keyUpDownUI(event)
		lsStreams.Items = strs
		parName.Text = Streams[StreamID].Status
		parGame.Text = Streams[StreamID].Game
		parViewers.Text = strconv.Itoa(Streams[StreamID].Viewers)
		parLength.Text = videoLen(Streams[StreamID].Length)
		termui.Render(termui.Body)
	})
	termui.Handle("/sys/kbd/<up>", func(event termui.Event) {
		strs = keyUpDownUI(event)
		lsStreams.Items = strs
		parName.Text = Streams[StreamID].Status
		parGame.Text = Streams[StreamID].Game
		parViewers.Text = strconv.Itoa(Streams[StreamID].Viewers)
		parLength.Text = videoLen(Streams[StreamID].Length)
		termui.Render(termui.Body)
	})
	termui.Handle("/sys/kbd/r", func(event termui.Event) {
		parNotiHelp.Text = "[Обновляю список стримов](fg-red)"
		termui.Render(parNotiHelp)
		parNotiHelp.Text = helpText
		Streams = getStreams(StreamFol, dataBase)
		StreamID = 0
		var strs []string
		for id, stream := range Streams {
			if id == 0 {
				strs = append(strs, "["+stream.DisplayName+"](fg-white,bg-green)")
			} else {
				strs = append(strs, stream.DisplayName)
			}
		}
		lsStreams.Items = strs
		parName.Text = Streams[StreamID].Status
		parGame.Text = Streams[StreamID].Game
		parViewers.Text = strconv.Itoa(Streams[StreamID].Viewers)
		parLength.Text = videoLen(Streams[StreamID].Length)
		termui.Render(termui.Body)
	})
	termui.Handle("/sys/kbd/o", func(event termui.Event) {
		parNotiHelp.Text = "[Обновляю список стримов](fg-red)"
		termui.Render(parNotiHelp)
		parNotiHelp.Text = helpText
		StreamFol = false
		Streams = getStreams(StreamFol, dataBase)
		StreamID = 0
		var strs []string
		for id, stream := range Streams {
			if id == 0 {
				strs = append(strs, "["+stream.DisplayName+"](fg-white,bg-green)")
			} else {
				strs = append(strs, stream.DisplayName)
			}
		}
		lsStreams.Items = strs
		parName.Text = Streams[StreamID].Status
		parGame.Text = Streams[StreamID].Game
		parViewers.Text = strconv.Itoa(Streams[StreamID].Viewers)
		parLength.Text = videoLen(Streams[StreamID].Length)
		termui.Render(termui.Body)
	})
	termui.Handle("/sys/kbd/f", func(event termui.Event) {
		parNotiHelp.Text = "[Обновляю список стримов](fg-red)"
		termui.Render(parNotiHelp)
		parNotiHelp.Text = helpText
		StreamFol = true
		Streams = getStreams(StreamFol, dataBase)
		StreamID = 0
		var strs []string
		for id, stream := range Streams {
			if id == 0 {
				strs = append(strs, "["+stream.DisplayName+"](fg-white,bg-green)")
			} else {
				strs = append(strs, stream.DisplayName)
			}
		}
		lsStreams.Items = strs
		parName.Text = Streams[StreamID].Status
		parGame.Text = Streams[StreamID].Game
		parViewers.Text = strconv.Itoa(Streams[StreamID].Viewers)
		parLength.Text = videoLen(Streams[StreamID].Length)
		termui.Render(termui.Body)
	})
	termui.Handle("/sys/kbd/<enter>", func(event termui.Event) {
		parNotiHelp.Text = "[Запускаю streamlink](fg-red)"
		termui.Render(parNotiHelp)
		cmd = exec.Command("streamlink", "-p", "mpv --fs", "--default-stream", "720p,720p60,best,source", Streams[StreamID].URL)
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			cmd.Wait()
			f, _ := os.Create("out")
			f.Write(out.Bytes())
			parNotiHelp.Text = helpText
			termui.Render(parNotiHelp)
		}()
	})
	termui.Handle("/sys/timer/5m", func(event termui.Event) {
		parNotiHelp.Text = "[Обновляю список стримов](fg-red)"
		termui.Render(parNotiHelp)
		parNotiHelp.Text = helpText
		Streams = getStreams(StreamFol, dataBase)
		StreamID = 0
		var strs []string
		for id, stream := range Streams {
			if id == 0 {
				strs = append(strs, "["+stream.DisplayName+"](fg-white,bg-green)")
			} else {
				strs = append(strs, stream.DisplayName)
			}
		}
		lsStreams.Items = strs
		parName.Text = Streams[StreamID].Status
		parGame.Text = Streams[StreamID].Game
		parViewers.Text = strconv.Itoa(Streams[StreamID].Viewers)
		parLength.Text = videoLen(Streams[StreamID].Length)
		termui.Render(termui.Body)
	})
	termui.Handle("/sys/wnd/resize", func(event termui.Event) {
		termui.Body.Width = termui.TermWidth()
		termui.Body.Align()
		termui.Clear()
		termui.Render(termui.Body)
	})
	termui.Loop()
}

func keyUpDownUI(event termui.Event) []string {
	streamCount := len(Streams) - 1
	if event.Path == "/sys/kbd/<down>" {
		if StreamID == streamCount {
			StreamID = 0
		} else {
			StreamID += 1
		}
	}
	if event.Path == "/sys/kbd/<up>" {
		if StreamID == 0 {
			StreamID = streamCount
		} else {
			StreamID -= 1
		}
	}

	var strs []string
	for id, stream := range Streams {
		if id == StreamID {
			strs = append(strs, "["+stream.DisplayName+"](fg-white,bg-green)")
		} else {
			strs = append(strs, stream.DisplayName)
		}
	}
	return strs
}

// TODO: err ->
func getStreams(streamsFol bool, dataBase DB) (streams []Stream) {
	tw := TWInit(clientID, redirectURI)
	if streamsFol == false {
		streams = tw.GetLive()
	} else {
		accessTokenRow, err := dataBase.SelectAccessToken()
		if err != nil {
			log.Panic(err)
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

			openbrowser(u.String())

			l, err := StartHttpServer()
			if err != nil {
				panic(err)
			}
			for ShutdownServer == false {
				time.Sleep(1 * time.Second)
			}
			l.Close()

			//if err := srv.Shutdown(nil); err != nil {
			//	fmt.Println(err)
			//}
			accessTokenRow, err = dataBase.SelectAccessToken()
			if err != nil {
				log.Panic(err)
			}
		}
		streams = tw.GetOnline(accessTokenRow.accessToken)
	}
	return streams
}

func openbrowser(url string) {
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
