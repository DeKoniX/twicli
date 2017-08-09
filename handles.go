package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gizak/termui"
)

func (app *Application) quitHandle(event termui.Event) {
	if app.Cmd != nil {
		app.Cmd.Process.Kill()
		app.Cmd = nil
	} else {
		termui.StopLoop()
	}
}

func (app *Application) upDownHandle(event termui.Event) {
	var streamCount = len(app.Streams) - 1

	switch event.Path {
	case "/sys/kbd/<down>":
		app.StreamID += 1
		if app.StreamID > streamCount {
			app.StreamID = 0
		}
	case "/sys/kbd/<up>":
		app.StreamID -= 1
		if app.StreamID < 0 {
			app.StreamID = streamCount
		}
	}
	app.updateStreamList(false, "")
}

func (app *Application) leftRightHandle(event termui.Event) {
	var str string

	var streamTypeCount = 2
	switch event.Path {
	case "/sys/kbd/<right>":
		app.StreamType += 1

		if app.StreamType > streamTypeCount {
			app.StreamType = 0
		}
	case "/sys/kbd/<left>":
		app.StreamType -= 1

		if app.StreamType < 0 {
			app.StreamType = streamTypeCount
		}
	}
	strs := [4]string{
		"<Ваши подписки>",
		"<Топ Twitch>",
		"<Топ RU Twitch>",
		"<Поиск>",
	}

	for id, s := range strs {
		if id == app.StreamType {
			s = "[" + s + "](bg-blue)"
		}
		str += s + " "
	}

	app.UI.parStreamType.Text = str
	termui.Render(app.UI.parStreamType)

	app.StreamID = 0
	app.updateStreamList(true, "")
}

func (app *Application) updateHandle(event termui.Event) {
	app.updateStreamList(true, "")
}

func (app *Application) runHandle(event termui.Event) {
	if len(app.Streams) > 0 {
		if app.Cmd != nil {
			app.Cmd.Process.Kill()
			app.Cmd = nil
			time.Sleep(10 * time.Millisecond)
		}
		app.UI.parNotiHelp.Text = "[Запускаю streamlink](fg-red)"
		termui.Render(app.UI.parNotiHelp)
		app.Cmd = exec.Command("streamlink", "-p", "mpv --fs", "--default-stream", "720p,720p60,best,source", app.Streams[app.StreamID].URL)
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
			app.UI.parNotiHelp.Text = helpText
			termui.Render(app.UI.parNotiHelp)
		}()
	}
}

func (app *Application) searchHandle(event termui.Event) {
	if event.Path == "/sys/kbd//" {
		app.UI.parNotiHelp.Text = ""
		termui.Render(app.UI.parNotiHelp)

		myHandlers := make(map[string]func(termui.Event))
		for path, handle := range termui.DefaultEvtStream.Handlers {
			myHandlers[path] = handle
		}
		termui.ResetHandlers()

		termui.Handle("/sys/kbd", func(event2 termui.Event) {
			keys := strings.Split(event2.Path, "/")
			app.UI.parNotiHelp.Text = app.UI.parNotiHelp.Text + keys[len(keys)-1]
			termui.Render(app.UI.parNotiHelp)
		})

		termui.Handle("/sys/kbd/<space>", func(event2 termui.Event) {
			app.UI.parNotiHelp.Text = app.UI.parNotiHelp.Text + " "
			termui.Render(app.UI.parNotiHelp)
		})

		termui.Handle("/sys/kbd/C-8", func(event2 termui.Event) {
			app.UI.parNotiHelp.Text = app.UI.parNotiHelp.Text[0 : len(app.UI.parNotiHelp.Text)-1]
			termui.Render(app.UI.parNotiHelp)
		})

		termui.Handle("/sys/kbd/<enter>", func(event2 termui.Event) {
			app.StreamType = 3
			app.StreamID = 0
			search := app.UI.parNotiHelp.Text
			app.updateStreamList(true, app.UI.parNotiHelp.Text)
			app.UI.parNotiHelp.Text = helpText + "\n Поиск: [" + search + "](fg-blue)"
			app.UI.parStreamType.Text = "<Ваши подписки> <Топ Twitch> <Топ RU Twitch> [<Поиск>](bg-blue)"

			termui.Render(termui.Body)

			termui.ResetHandlers()
			for path, handle := range myHandlers {
				termui.Handle(path, handle)
			}
		})
	}
}
