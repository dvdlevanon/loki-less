package ui

import (
	"fmt"
	"time"

	"github.com/dvdlevanon/loki-less/pkg/logstream"
	"github.com/gdamore/tcell"
)

func NewMainWindow(stream *logstream.LogStream) (*MainWindow, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}

	screen.EnableMouse()
	screen.Clear()

	logView := NewLogView(screen, stream)
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
		width, _ := w.screen.Size()
		drawText(w.screen, width-100, 1, width, 2, tcell.StyleDefault, fmt.Sprintf("%v", event.Name()))

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
	width, height := w.screen.Size()
	drawBox(w.screen, 0, 0, width-1, height-1, tcell.StyleDefault)
	w.logView.refresh()
	w.screen.Show()
}

func (w *MainWindow) resize() {
	w.screen.Sync()
	width, height := w.screen.Size()
	w.logView.viewport.setRows(height - 2)
	w.logView.setSize(1, 1, width-1, height-2)
}

func (w *MainWindow) pollEvents() {
	for {
		event := w.screen.PollEvent()
		w.events <- event
	}
}
