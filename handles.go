package main

import (
	"strconv"
	"strings"

	"github.com/gizak/termui"
)

func (app *Application) quitHandle(event termui.Event) {
	if app.Cmd != nil {
		app.Cmd.Process.Kill()
		app.Cmd = nil
		app.UI.parStreamOn.Text = ""
		app.StreamNowName = ""
		termui.Render(app.UI.parStreamOn)
	} else {
		termui.StopLoop()
	}
}

func (app *Application) upDownHandle(event termui.Event) {
	var streamCount = len(app.Streams) - 1

	switch event.Path {
	case "/sys/kbd/<down>":
		app.StreamID++
		if app.StreamID > streamCount {
			app.StreamID = 0
		}
	case "/sys/kbd/<up>":
		app.StreamID--
		if app.StreamID < 0 {
			app.StreamID = streamCount
		}
	}
	app.updateStreamList(false, "")
}

func (app *Application) pageUpPageDownHandle(event termui.Event) {
	app.StreamID = 0
	switch event.Path {
	case "/sys/kbd/<next>":
		app.StreamPage++
	case "/sys/kbd/<previous>":
		if app.StreamPage != 0 {
			app.StreamPage--
		}
	}
	app.UI.parPageStream.Text = "[" + strconv.Itoa(app.StreamPage+1) + "](fg-green)"
	app.updateStreamList(true, app.Search)
}

func (app *Application) leftRightHandle(event termui.Event) {
	var str string

	var streamTypeCount = 3
	switch event.Path {
	case "/sys/kbd/<right>":
		app.StreamType++

		if app.StreamType > streamTypeCount-1 && app.Search == "" || app.StreamType > streamTypeCount {
			app.StreamType = 0
		}
	case "/sys/kbd/<left>":
		app.StreamType--

		if app.StreamType < 0 {
			if app.Search != "" {
				app.StreamType = streamTypeCount
			} else {
				app.StreamType = streamTypeCount - 1
			}
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

	app.StreamPage = 0
	app.UI.parPageStream.Text = "[" + strconv.Itoa(app.StreamPage+1) + "](fg-green)"

	app.StreamID = 0
	if app.StreamType == 3 {
		app.updateStreamList(true, app.Search)
	} else {
		app.updateStreamList(true, "")
	}
}

func (app *Application) updateHandle(event termui.Event) {
	app.updateStreamList(true, app.Search)
}

func (app *Application) runHandle(event termui.Event) {
	app.runStreamlink()
}

func (app *Application) searchHandle(event termui.Event) {
	if event.Path == "/sys/kbd//" {
		app.Search = ""
		app.UI.parNotiHelp.Text = "Поиск:"
		termui.Render(app.UI.parNotiHelp)

		myHandlers := make(map[string]func(termui.Event))
		for path, handle := range termui.DefaultEvtStream.Handlers {
			myHandlers[path] = handle
		}
		termui.ResetHandlers()

		termui.Handle("/sys/kbd", func(event2 termui.Event) {
			keys := strings.Split(event2.Path, "/")
			app.Search += keys[len(keys)-1]
			app.UI.parNotiHelp.Text = "Поиск: " + app.Search
			termui.Render(app.UI.parNotiHelp)
		})

		termui.Handle("/sys/kbd/<space>", func(event2 termui.Event) {
			app.Search += " "
			app.UI.parNotiHelp.Text = "Поиск: " + app.Search
			termui.Render(app.UI.parNotiHelp)
		})

		termui.Handle("/sys/kbd/C-8", func(event2 termui.Event) {
			runes := []rune(app.Search)
			app.Search = string(runes[:len(runes)-1])
			app.UI.parNotiHelp.Text = "Поиск: " + app.Search
			termui.Render(app.UI.parNotiHelp)
		})

		termui.Handle("/sys/kbd/<escape>", func(event2 termui.Event) {
			app.UI.parNotiHelp.Text = helpText
			termui.Render(app.UI.parNotiHelp)
			termui.ResetHandlers()
			for path, handle := range myHandlers {
				termui.Handle(path, handle)
			}
		})

		termui.Handle("/sys/kbd/<enter>", func(event2 termui.Event) {
			app.StreamPage = 0
			app.UI.parPageStream.Text = "[" + strconv.Itoa(app.StreamPage+1) + "](fg-green)"
			app.StreamType = 3
			app.StreamID = 0
			app.updateStreamList(true, app.Search)
			app.UI.parNotiHelp.Text = helpText + "\n Поиск: [" + app.Search + "](fg-blue)"
			app.UI.parStreamType.Text = "<Ваши подписки> <Топ Twitch> <Топ RU Twitch> [<Поиск>](bg-blue)"

			termui.Render(termui.Body)

			termui.ResetHandlers()
			for path, handle := range myHandlers {
				termui.Handle(path, handle)
			}
		})
	}
}
