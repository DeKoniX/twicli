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

	"strings"

	"github.com/gizak/termui"
)

const clientID = ""
const redirectURI = "http://localhost:5454"

const helpText = "[<up>, <down>] - вверх, вниз по списку доступных стримов [r] - обновить список стримов [q] - выйти из приложения\n [<right>, <left>] - бегать по вкладкам приложения [/] - поиск по Twitch [enter] - запустить streamlink"

// Type: 0 - sub, 1 - top, 2 - ru top, 3 - search

func main() {
	var Streams []Stream
	var StreamID int = 0
	var StreamType int = 0

	var err error
	var cmd *exec.Cmd

	dataBase, err := initDB()
	if err != nil {
		log.Panic(err)
	}
	Streams = getStreams(StreamType, dataBase, "")
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

	parStreamType := termui.NewPar("[<Ваши подписки>](bg-blue) <Топ Twitch> <Топ RU Twitch> <Поиск>")
	parStreamType.Height = 1
	parStreamType.Border = false

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
			termui.NewCol(10, 2, parStreamType),
		),
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
		strs, StreamID = keyUpDownUI(event, Streams, StreamID)
		lsStreams.Items = strs
		if len(Streams) > 0 {
			parName.Text = Streams[StreamID].Status
			parGame.Text = Streams[StreamID].Game
			parViewers.Text = strconv.Itoa(Streams[StreamID].Viewers)
			parLength.Text = videoLen(Streams[StreamID].Length)
		}
		termui.Render(termui.Body)
	})
	termui.Handle("/sys/kbd/<up>", func(event termui.Event) {
		strs, StreamID = keyUpDownUI(event, Streams, StreamID)
		lsStreams.Items = strs
		if len(Streams) > 0 {
			parName.Text = Streams[StreamID].Status
			parGame.Text = Streams[StreamID].Game
			parViewers.Text = strconv.Itoa(Streams[StreamID].Viewers)
			parLength.Text = videoLen(Streams[StreamID].Length)
		}
		termui.Render(termui.Body)
	})
	termui.Handle("/sys/kbd/<right>", func(event termui.Event) {
		var strs []string
		var str string
		str, StreamType = keyLeftRightUI(event, StreamType)
		parStreamType.Text = str
		parNotiHelp.Text = "[Обновляю список стримов](fg-red)"
		termui.Render(parNotiHelp)
		termui.Render(parStreamType)
		time.Sleep(10 * time.Millisecond)
		parNotiHelp.Text = helpText
		Streams = getStreams(StreamType, dataBase, "")
		StreamID = 0
		for id, stream := range Streams {
			if id == 0 {
				strs = append(strs, "["+stream.DisplayName+"](fg-white,bg-green)")
			} else {
				strs = append(strs, stream.DisplayName)
			}
		}
		lsStreams.Items = strs
		if len(Streams) > 0 {
			parName.Text = Streams[StreamID].Status
			parGame.Text = Streams[StreamID].Game
			parViewers.Text = strconv.Itoa(Streams[StreamID].Viewers)
			parLength.Text = videoLen(Streams[StreamID].Length)
		}
		termui.Render(termui.Body)
	})
	termui.Handle("/sys/kbd/<left>", func(event termui.Event) {
		var strs []string
		var str string
		str, StreamType = keyLeftRightUI(event, StreamType)
		parStreamType.Text = str
		parNotiHelp.Text = "[Обновляю список стримов](fg-red)"
		termui.Render(parNotiHelp)
		termui.Render(parStreamType)
		time.Sleep(10 * time.Millisecond)
		parNotiHelp.Text = helpText
		Streams = getStreams(StreamType, dataBase, "")
		StreamID = 0
		for id, stream := range Streams {
			if id == 0 {
				strs = append(strs, "["+stream.DisplayName+"](fg-white,bg-green)")
			} else {
				strs = append(strs, stream.DisplayName)
			}
		}
		lsStreams.Items = strs
		if len(Streams) > 0 {
			parName.Text = Streams[StreamID].Status
			parGame.Text = Streams[StreamID].Game
			parViewers.Text = strconv.Itoa(Streams[StreamID].Viewers)
			parLength.Text = videoLen(Streams[StreamID].Length)
		}
		termui.Render(termui.Body)
	})
	termui.Handle("/sys/kbd/r", func(event termui.Event) {
		parNotiHelp.Text = "[Обновляю список стримов](fg-red)"
		termui.Render(parNotiHelp)
		time.Sleep(10 * time.Millisecond)
		parNotiHelp.Text = helpText
		Streams = getStreams(StreamType, dataBase, "")
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
		if len(Streams) > 0 {
			parName.Text = Streams[StreamID].Status
			parGame.Text = Streams[StreamID].Game
			parViewers.Text = strconv.Itoa(Streams[StreamID].Viewers)
			parLength.Text = videoLen(Streams[StreamID].Length)
		}
		termui.Render(termui.Body)
	})
	termui.Handle("/sys/kbd/<enter>", func(event termui.Event) {
		if len(Streams) > 0 {
			if cmd != nil {
				cmd.Process.Kill()
				cmd = nil
				time.Sleep(10 * time.Millisecond)
			}
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
		}
	})
	termui.Handle("/sys/kbd//", func(event termui.Event) {
		if event.Path == "/sys/kbd//" {
			parNotiHelp.Text = ""
			termui.Render(parNotiHelp)

			myHandlers := make(map[string]func(termui.Event))
			for path, handle := range termui.DefaultEvtStream.Handlers {
				myHandlers[path] = handle
			}
			termui.ResetHandlers()

			termui.Handle("/sys/kbd", func(event2 termui.Event) {
				keys := strings.Split(event2.Path, "/")
				parNotiHelp.Text = parNotiHelp.Text + keys[len(keys)-1]
				termui.Render(parNotiHelp)
			})

			termui.Handle("/sys/kbd/<space>", func(event2 termui.Event) {
				parNotiHelp.Text = parNotiHelp.Text + " "
				termui.Render(parNotiHelp)
			})

			termui.Handle("/sys/kbd/C-8", func(event2 termui.Event) {
				parNotiHelp.Text = parNotiHelp.Text[0 : len(parNotiHelp.Text)-1]
				termui.Render(parNotiHelp)
			})

			termui.Handle("/sys/kbd/<enter>", func(event2 termui.Event) {
				StreamType = 3
				Streams = getStreams(StreamType, dataBase, parNotiHelp.Text)
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
				if len(Streams) > 0 {
					parName.Text = Streams[StreamID].Status
					parGame.Text = Streams[StreamID].Game
					parViewers.Text = strconv.Itoa(Streams[StreamID].Viewers)
					parLength.Text = videoLen(Streams[StreamID].Length)
				}
				parNotiHelp.Text = helpText + "\n Поиск: [" + parNotiHelp.Text + "](fg-blue)"
				parStreamType.Text = "<Ваши подписки> <Топ Twitch> <Топ RU Twitch> [<Поиск>](bg-blue)"

				termui.Render(termui.Body)

				termui.ResetHandlers()
				for path, handle := range myHandlers {
					termui.Handle(path, handle)
				}
			})
		}
	})
	termui.Handle("/sys/timer/5m", func(event termui.Event) {
		parNotiHelp.Text = "[Обновляю список стримов](fg-red)"
		termui.Render(parNotiHelp)
		parNotiHelp.Text = helpText
		Streams = getStreams(StreamType, dataBase, "")
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

func keyUpDownUI(event termui.Event, streams []Stream, streamID int) (strs []string, _ int) {
	streamCount := len(streams) - 1
	if event.Path == "/sys/kbd/<down>" {
		if streamID == streamCount {
			streamID = 0
		} else {
			streamID += 1
		}
	}
	if event.Path == "/sys/kbd/<up>" {
		if streamID == 0 {
			streamID = streamCount
		} else {
			streamID -= 1
		}
	}

	for id, stream := range streams {
		if id == streamID {
			strs = append(strs, "["+stream.DisplayName+"](fg-white,bg-green)")
		} else {
			strs = append(strs, stream.DisplayName)
		}
	}
	return strs, streamID
}

func keyLeftRightUI(event termui.Event, streamType int) (str string, _ int) {
	var streamTypeCount = 2

	switch event.Path {
	case "/sys/kbd/<right>":
		streamType += 1

		if streamType > streamTypeCount {
			streamType = 0
		}
	case "/sys/kbd/<left>":
		streamType -= 1

		if streamType < 0 {
			streamType = streamTypeCount
		}
	}

	strs := [4]string{
		"<Ваши подписки>",
		"<Топ Twitch>",
		"<Топ RU Twitch>",
		"<Поиск>",
	}

	for id, s := range strs {
		if id == streamType {
			s = "[" + s + "](bg-blue)"
		}
		str += s + " "
	}

	return str, streamType
}

// TODO: err ->
func getStreams(streamsType int, dataBase DB, search string) (streams []Stream) {
	tw := TWInit(clientID, redirectURI)

	switch streamsType {
	case 1:
		streams = tw.GetLive("")
	case 2:
		streams = tw.GetLive("ru")
	case 3:
		streams = tw.GetSearch(search)
	case 0:
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

			openBrowser(u.String())

			l, err := StartHttpServer()
			if err != nil {
				panic(err)
			}
			for ShutdownServer == false {
				time.Sleep(1 * time.Second)
			}
			l.Close()
			accessTokenRow, err = dataBase.SelectAccessToken()
			if err != nil {
				log.Panic(err)
			}
		}
		streams = tw.GetOnline(accessTokenRow.accessToken)
	}
	return streams
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
