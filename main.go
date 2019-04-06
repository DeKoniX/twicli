package main

import (
	"os/exec"
	"time"

	"strconv"

	"fmt"
	"os"

	"github.com/gizak/termui"
)

const clientID = "jcmupzvav5wezhdo0phjwcsvkc6t00"
const redirectURI = "http://localhost:5454"

const helpText = "[<up>, <down>] - вверх, вниз по списку доступных стримов [pageup, pagedown] - выбор страницы списка стримов\n[<right>, <left>] - бегать по вкладкам приложения [r] - обновить список стримов [q] - выйти из приложения\n[/] - поиск по Twitch [enter] - запустить streamlink [\\] - запустить streamlink(c выбором качества стрима)"

// Type: 0 - sub, 1 - top, 2 - ru top, 3 - search, 4 - get quality

type UIWidgets struct {
	parPageStream *termui.Par
	parStreamOn   *termui.Par
	parStreamType *termui.Par
	lsStreams     *termui.List
	parName       *termui.Par
	parGame       *termui.Par
	parViewers    *termui.Par
	parLength     *termui.Par
	parNotiHelp   *termui.Par
}

type Application struct {
	Cmd           *exec.Cmd
	DB            DB
	QualityID     int
	Search        string
	StreamID      int
	StreamNowName string
	StreamPage    int
	StreamType    int
	Streams       []Stream
	TW            *TW
	UI            UIWidgets
}

func main() {
	var app Application
	var err error

	app.StreamID = 0
	app.StreamType = 0
	app.StreamPage = 0

	app.DB, err = initDB()
	if err != nil {
		fmt.Println("Ошибка инициализации базы данных: ", err)
		os.Exit(2)
	}
	app.TW = TWInit(clientID, redirectURI)

	app.Streams, err = app.getStreams(app.StreamType, app.DB, "", app.StreamPage)
	if err != nil {
		fmt.Println("Ошибка получения списка стримов: ", err)
		os.Exit(2)
	}

	var strs []string
	for id, stream := range app.Streams {
		if id == 0 {
			strs = append(strs, "["+stream.DisplayName+"](fg-white,bg-green)")
		} else {
			strs = append(strs, stream.DisplayName)
		}
	}

	err = termui.Init()
	if err != nil {
		fmt.Println("Ошибка инициализации termui: ", err)
		os.Exit(2)
	}
	defer termui.Close()

	app.UI.parPageStream = termui.NewPar("[1](fg-green)")
	app.UI.parPageStream.Height = 1
	app.UI.parPageStream.Border = false

	app.UI.parStreamOn = termui.NewPar("")
	app.UI.parStreamOn.Height = 1
	app.UI.parStreamOn.Border = false

	app.UI.parStreamType = termui.NewPar("[<Ваши подписки>](bg-blue) <Топ Twitch> <Топ RU Twitch> <Поиск>")
	app.UI.parStreamType.Height = 1
	app.UI.parStreamType.Border = false

	app.UI.lsStreams = termui.NewList()
	app.UI.lsStreams.Items = strs
	app.UI.lsStreams.BorderLabel = "Стримы:"
	app.UI.lsStreams.Height = 12

	app.UI.parName = termui.NewPar(app.Streams[0].Status)
	app.UI.parName.Height = 3
	app.UI.parName.BorderLabel = "Наименование:"

	app.UI.parGame = termui.NewPar(app.Streams[0].Game)
	app.UI.parGame.Height = 3
	app.UI.parGame.BorderLabel = "Игра:"

	app.UI.parViewers = termui.NewPar(strconv.Itoa(app.Streams[0].Viewers))
	app.UI.parViewers.Height = 3
	app.UI.parViewers.BorderLabel = "Смотрят:"

	app.UI.parLength = termui.NewPar(videoLen(app.Streams[0].Length))
	app.UI.parLength.Height = 3
	app.UI.parLength.BorderLabel = "Идет уже:"

	app.UI.parNotiHelp = termui.NewPar(helpText)
	app.UI.parNotiHelp.Height = 4
	app.UI.parNotiHelp.Border = false

	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(2, 0, app.UI.parPageStream),
			termui.NewCol(1, 0, app.UI.parStreamOn),
			termui.NewCol(8, 1, app.UI.parStreamType),
		),
		termui.NewRow(
			termui.NewCol(3, 0, app.UI.lsStreams),
			termui.NewCol(9, 0, app.UI.parName, app.UI.parGame, app.UI.parViewers, app.UI.parLength),
		),
		termui.NewRow(
			termui.NewCol(12, 0, app.UI.parNotiHelp),
		),
	)

	termui.Body.Align()

	termui.Render(termui.Body)

	termui.Handle("/sys/kbd/q", app.quitHandle)
	termui.Handle("/sys/kbd/<down>", app.upDownHandle)
	termui.Handle("/sys/kbd/<up>", app.upDownHandle)
	termui.Handle("/sys/kbd/<next>", app.pageUpPageDownHandle)
	termui.Handle("/sys/kbd/<previous>", app.pageUpPageDownHandle)
	termui.Handle("/sys/kbd/<right>", app.leftRightHandle)
	termui.Handle("/sys/kbd/<left>", app.leftRightHandle)
	termui.Handle("/sys/kbd/r", app.updateHandle)
	termui.Handle("/sys/kbd/<enter>", app.runHandle)
	termui.Handle("/sys/kbd/\\", app.runQualityHandle)
	termui.Handle("/sys/kbd//", app.searchHandle)
	termui.Handle("/sys/wnd/resize", func(event termui.Event) {
		termui.Body.Width = termui.TermWidth()
		termui.Body.Align()
		termui.Clear()
		termui.Render(termui.Body)
	})

	termui.Merge("timer", termui.NewTimerCh(5*time.Minute))
	termui.Handle("/timer/5m", app.updateHandle)

	termui.Loop()
}
