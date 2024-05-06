package ui

import (
	"fmt"

	"github.com/dvdlevanon/loki-less/pkg/logstream"
	"github.com/gdamore/tcell"
)

func NewLogView(screen tcell.Screen, stream *logstream.LogStream, requests chan<- logstream.ChunkRequest) LogView {
	return LogView{
		screen:   screen,
		viewport: NewLogViewPort(stream, requests),
		requests: requests,
	}
}

type LogView struct {
	screen              tcell.Screen
	x, y, columns, rows int
	viewport            *LogViewPort
	requests            chan<- logstream.ChunkRequest
}

func (l *LogView) setSize(x, y, columns, rows int) {
	l.x = x
	l.y = y
	l.columns = columns
	l.rows = rows
	l.viewport.setRows(rows)
}

func (l *LogView) refresh() {
	row := l.y
	endRow := l.rows + l.y
	chunk, offset := l.viewport.refresh()
	prevChunk := chunk

	for chunk != nil {
		switch chunk.Type() {
		case logstream.RAM_CHUNK:
			prevChunk = chunk
			chunk, offset, row = l.handleRamChunk(chunk, offset, row, endRow)
		case logstream.LOADING_CHUNK:
			chunk, offset, row = l.handleLoading(chunk, row, endRow)
		default:
			return
		}
	}

	if prevChunk != nil && row < endRow {
		l.pushRequest(prevChunk.NextRequest())
	}
}

func (l *LogView) handleLoading(chunk *logstream.LogChunk, row int, endRow int) (*logstream.LogChunk, int, int) {
	if row > endRow {
		return nil, 0, row + 1
	}

	seconds := chunk.ElapsedLoadingTime() / 1000
	drawText(l.screen, 1, row, l.columns, endRow, tcell.StyleDefault, fmt.Sprintf("loading... (%d seconds)", seconds))
	return chunk.Next(), 0, row + 1
}

func (l *LogView) handleRamChunk(chunk *logstream.LogChunk, offset int, row int, endRow int) (*logstream.LogChunk, int, int) {
	rowsTotal := l.showChunk(chunk, offset, row, endRow)

	if row >= endRow {
		return nil, 0, row + rowsTotal
	}

	row = row + rowsTotal
	offset = offset + rowsTotal

	if offset >= len(chunk.Lines()) {
		return chunk.Next(), 0, row
	}

	return chunk, offset, row
}

func (l *LogView) showChunk(chunk *logstream.LogChunk, chunkOffset int, startRow int, endRow int) int {
	lineOffest := chunkOffset
	for i := startRow; i < endRow; i++ {
		if lineOffest >= len(chunk.Lines()) {
			return i - startRow
		}

		line := chunk.Lines()[lineOffest]
		drawText(l.screen, l.x, i, l.columns, i+1, tcell.StyleDefault, line.FormattedLine())
		lineOffest = lineOffest + 1
	}

	return endRow - 1
}

func (l *LogView) pushRequest(request *logstream.ChunkRequest) {
	if request == nil {
		return
	}

	l.requests <- *request
}
