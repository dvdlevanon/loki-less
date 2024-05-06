package ui

import (
	"time"

	"github.com/dvdlevanon/loki-less/pkg/logstream"
)

func NewLogViewPort(stream *logstream.LogStream, requests chan<- logstream.ChunkRequest) *LogViewPort {
	return &LogViewPort{
		stream:   stream,
		requests: requests,
	}
}

type LogViewPort struct {
	stream           *logstream.LogStream
	rows             int
	chunk            *logstream.LogChunk
	chunkOffset      int
	autoScroll       bool
	requests         chan<- logstream.ChunkRequest
	tryToScroll      int
	tryToScrollChunk *logstream.LogChunk
}

func (v *LogViewPort) setRows(rows int) {
	v.rows = rows
}

func (v *LogViewPort) refresh() (*logstream.LogChunk, int) {
	if v.chunk == nil {
		v.chunk = v.stream.Head()
		v.chunkOffset = 0
	}

	if v.chunk == nil {
		v.pushRequest(v.stream.TimeRequest(time.Now().UnixNano()), nil, 0)
		return nil, 0
	}

	v.refreshLoadingChunk()
	v.refreshTryToScroll()

	if v.autoScroll {
		v.ScrollToLatest()
		v.pushRequest(v.stream.Tail().NextRequest(), nil, 0)
	}

	return v.chunk, v.chunkOffset
}

func (v *LogViewPort) refreshLoadingChunk() {
	if v.chunk.Type() != logstream.LOADING_CHUNK {
		return
	}

	if v.chunk.LoadedChunk() != nil {
		if v.chunk.LoadingForward() {
			v.chunkOffset = 0
		} else {
			v.chunkOffset = v.chunk.LoadedChunk().LineCount() - 1
		}

		v.chunk = v.chunk.LoadedChunk()
		v.chunkOffset = 0
	}
}

func (v *LogViewPort) refreshTryToScroll() {
	chunk := v.tryToScrollChunk
	if v.tryToScroll == 0 || chunk == nil {
		return
	}

	if v.tryToScroll < 0 {
		chunk = chunk.Prev()
	} else {
		chunk = chunk.Next()
	}

	if chunk == nil {
		return
	}

	v.Scroll(v.tryToScroll)
	v.tryToScroll = 0
}

func (v *LogViewPort) AutoScrollOn() {
	v.autoScroll = true
}

func (v *LogViewPort) AutoScrollOff() {
	v.autoScroll = false
}

func (v *LogViewPort) ScrollToBeginning() {
	v.chunk = v.chunk.Stream().Head()
	v.chunkOffset = 0
}

func (v *LogViewPort) ScrollToLatest() {
	v.chunk = v.chunk.Stream().Tail()

	if v.chunk == nil {
		return
	}

	v.chunkOffset = v.chunk.LineCount()
	shownLines := 0

	for shownLines < v.rows {
		v.chunkOffset = v.chunkOffset - 1
		shownLines = shownLines + 1
		if v.chunkOffset <= 0 {
			if v.chunk.Prev() == nil {
				return
			}
			v.chunk = v.chunk.Prev()
			v.chunkOffset = v.chunk.LineCount()
		}
	}
}

func (w *LogViewPort) Scroll(offset int) {
	logger.Debugf("Scrolling %d", offset)

	inChunkOffset := w.scrollChunks(w.chunkOffset + offset)
	w.chunkOffset = w.scrollInChunk(offset > 0, inChunkOffset)
}

func (w *LogViewPort) scrollChunks(offset int) int {
	shouldContinue := true
	for {
		if !shouldContinue || w.chunk == nil {
			return offset
		} else if offset < 0 {
			offset = w.scrollToPrevChunk(offset)
		} else if offset != 0 && offset >= w.chunk.LineCount() {
			offset, shouldContinue = w.scrollToNextChunk(offset)
		} else {
			return offset
		}
	}
}

func (w *LogViewPort) scrollToPrevChunk(offset int) int {
	if w.chunk.Prev() == nil {
		w.pushRequest(w.chunk.PrevRequest(), w.chunk, -1)
		return 0
	}

	w.chunk = w.chunk.Prev()

	if !w.chunk.Viewable() {
		return 0
	}

	return w.chunk.LineCount() + offset
}

func (w *LogViewPort) scrollToNextChunk(offset int) (int, bool) {
	if w.chunk.Next() == nil {
		w.pushRequest(w.chunk.NextRequest(), w.chunk, 1)
		return 0, true
	}

	if w.isViewportEmptied(w.chunk, w.chunk.Next(), w.chunkOffset, 0) {
		return w.chunk.LineCount() - 1, false
	}

	newOffset := offset - w.chunk.LineCount()
	w.chunk = w.chunk.Next()
	w.chunkOffset = 0
	return newOffset, true
}

func (w *LogViewPort) scrollInChunk(scrollingForward bool, offsetInChunk int) int {
	if !scrollingForward {
		return offsetInChunk
	}

	if w.isViewportEmptied(w.chunk, w.chunk, w.chunkOffset, offsetInChunk) {
		w.requestLastChunk(w.chunk)
		offsetInChunk = w.lastPossibleOffset(w.chunk, w.chunk, w.chunkOffset, offsetInChunk)
	}

	return offsetInChunk
}

func (w *LogViewPort) requestLastChunk(chunk *logstream.LogChunk) {
	if chunk == nil {
		return
	}

	for chunk.Next() != nil {
		chunk = chunk.Next()
	}

	if chunk.Type() != logstream.LOADING_CHUNK {
		w.pushRequest(chunk.NextRequest(), chunk, 1)
	}
}

func (w *LogViewPort) lastPossibleOffset(beforeChunk, afterChunk *logstream.LogChunk, beforeOffset, afterOffset int) int {
	for afterOffset > 0 && w.isViewportEmptied(beforeChunk, afterChunk, beforeOffset, afterOffset) {
		afterOffset -= 1
	}

	return afterOffset
}

func (w *LogViewPort) isViewportEmptied(beforeChunk, afterChunk *logstream.LogChunk, beforeOffset, afterOffset int) bool {
	afterRows := w.viewableRows(afterChunk, afterOffset)
	if afterRows >= w.rows {
		return false
	}

	beforeRows := w.viewableRows(beforeChunk, beforeOffset)
	return afterRows < beforeRows
}

func (v *LogViewPort) viewableRows(chunk *logstream.LogChunk, offset int) int {
	viewableRows := 0
	cur := chunk

	for cur != nil && viewableRows < v.rows {
		if cur.LineCount() < offset {
			viewableRows += cur.LineCount()
			offset -= cur.LineCount()
		} else {
			viewableRows += (cur.LineCount() - offset)
			offset = 0
		}

		cur = cur.Next()
	}

	return viewableRows
}

func (w *LogViewPort) pushRequest(request *logstream.ChunkRequest, chunk *logstream.LogChunk, tryToScroll int) {
	if request == nil {
		return
	}

	w.tryToScrollChunk = chunk
	w.tryToScroll = tryToScroll
	w.requests <- *request
}
