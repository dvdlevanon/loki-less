package ui

import (
	"github.com/dvdlevanon/loki-less/pkg/logstream"
	"github.com/gdamore/tcell"
)

type LogView struct {
	stream              *logstream.LogStream
	screen              tcell.Screen
	x, y, columns, rows int
	chunk               *logstream.LogChunk
	chunkOffset         int
	autoScroll          bool
}

func (l *LogView) autoScrollOn() {
	l.autoScroll = true
}

func (l *LogView) autoScrollOff() {
	l.autoScroll = false
}

func (l *LogView) setSize(x, y, columns, rows int) {
	l.x = x
	l.y = y
	l.columns = columns
	l.rows = rows
}

func (l *LogView) refresh() {
	if l.chunk == nil {
		l.chunk = l.stream.Head()
		l.chunkOffset = 0
	}

	if l.chunk == nil {
		return
	}

	if l.autoScroll {
		l.scrollToLatest()
	}

	row := l.y
	endRow := l.rows + l.y
	curChunk := l.chunk
	curOffset := l.chunkOffset

	for {
		if curChunk == nil || curChunk.Type() != logstream.RAM_CHUNK {
			break
		}

		rowsTotal := l.showChunk(curChunk, curOffset, row)

		if row >= endRow {
			break
		}

		row = row + rowsTotal
		curOffset = curOffset + rowsTotal

		if curOffset >= len(curChunk.Lines()) {
			curChunk = curChunk.Next()
			curOffset = 0
		}
	}
}

func (l *LogView) showChunk(chunk *logstream.LogChunk, chunkOffset int, startRow int) int {
	lineOffest := chunkOffset
	for i := startRow; i < l.rows; i++ {
		if lineOffest >= len(chunk.Lines()) {
			return i - startRow
		}

		line := chunk.Lines()[lineOffest]
		drawText(l.screen, l.x, i, l.columns, i+1, tcell.StyleDefault, line.FormattedLine())
		lineOffest = lineOffest + 1
	}

	return l.rows
}

func (l *LogView) scrollToBeginning() {
	l.chunk = l.stream.Head()
	l.chunkOffset = 0
}

func (l *LogView) scrollToLatest() {
	l.chunk = l.stream.Tail()
	l.chunkOffset = l.chunk.LineCount()
	shownLines := 0

	for shownLines < l.rows {
		l.chunkOffset = l.chunkOffset - 1
		shownLines = shownLines + 1
		if l.chunkOffset < 0 {
			l.chunk = l.chunk.Prev()
			l.chunkOffset = l.chunk.LineCount()
		}
	}
}

func (l *LogView) scroll(offset int) {
	newOffset := l.chunkOffset + offset

	for {
		if newOffset < 0 {
			if l.chunk.Prev() == nil {
				return
			}

			l.chunk = l.chunk.Prev()
			newOffset = l.chunk.LineCount() + newOffset
		} else if newOffset > l.chunk.LineCount() {
			if l.chunk.Next() == nil {
				return
			}
			newOffset = newOffset - l.chunk.LineCount()
			l.chunk = l.chunk.Next()
		} else {
			break
		}
	}

	l.chunkOffset = newOffset
}
