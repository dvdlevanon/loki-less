package ui

import (
	"time"

	"github.com/dvdlevanon/loki-less/pkg/logstream"
	"github.com/gdamore/tcell"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("ui")

func NewMainWindow(stream *logstream.LogStream, requests chan<- logstream.ChunkRequest) (*MainWindow, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}

	screen.EnableMouse()
	screen.Clear()

	logView := NewLogView(screen, stream, requests)
	result := &MainWindow{
		screen:  screen,
		logView: logView,
		stream:  stream,
		events:  make(chan tcell.Event),
	}
	result.resize()

	return result, nil
}

type MainWindow struct {
	stream  *logstream.LogStream
	screen  tcell.Screen
	logView LogView
	events  chan tcell.Event
}

func (w *MainWindow) Show() {
	ticker := time.NewTicker(time.Second / 100)
	defer ticker.Stop()

	go w.pollEvents()
	for {
		select {
		case <-ticker.C:
			w.refresh()
		case event := <-w.events:
			w.screen.Clear()
			if !w.handleEvent(event) {
				return
			}
			w.refresh()
		}
	}
}

func (w *MainWindow) Release() {
	w.screen.Fini()
}

func (w *MainWindow) handleEvent(event tcell.Event) bool {
	switch event := event.(type) {
	case *tcell.EventResize:
		w.resize()
	case *tcell.EventKey:
		if event.Key() == tcell.KeyRune {
			if event.Rune() == 'q' {
				return false
			}

			if event.Rune() == 'F' {
				w.logView.viewport.AutoScrollOn()
			}
		}

		if event.Key() == tcell.KeyCtrlC {
			w.logView.viewport.AutoScrollOff()
		}

		if event.Key() == tcell.KeyDown {
			w.logView.viewport.Scroll(1)
		}

		if event.Key() == tcell.KeyUp {
			w.logView.viewport.Scroll(-1)
		}

		if event.Key() == tcell.KeyPgUp {
			w.logView.viewport.Scroll(-50)
		}

		if event.Key() == tcell.KeyPgDn {
			w.logView.viewport.Scroll(50)
		}

		if event.Key() == tcell.KeyEnd {
			w.logView.viewport.ScrollToLatest()
		}

		if event.Key() == tcell.KeyHome {
			w.logView.viewport.ScrollToBeginning()
		}
	}

	return true
}

func (w *MainWindow) refresh() {
	startMillis := time.Now().UnixMilli()
	width, height := w.screen.Size()
	drawBox(w.screen, 0, 0, width-1, height-1, tcell.StyleDefault)
	w.logView.refresh()
	w.screen.Show()
	doneMillis := time.Now().UnixMilli() - startMillis
	if doneMillis > 1000 {
		logger.Warningf("Long refresh %d", doneMillis)
	}
}

func (w *MainWindow) resize() {
	w.screen.Sync()
	width, height := w.screen.Size()
	header := 5
	footer := 1
	w.logView.setSize(1, header, width-1, height-header-footer)
}

func (w *MainWindow) pollEvents() {
	for {
		event := w.screen.PollEvent()
		w.events <- event
	}
}
