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
	newOffset := w.chunkOffset + offset

	shouldContinue := true
	for {
		if !shouldContinue || w.chunk == nil {
			return
		} else if newOffset < 0 {
			newOffset = w.scrollToPrevChunk(newOffset)
		} else if newOffset != 0 && newOffset >= w.chunk.LineCount() {
			newOffset, shouldContinue = w.scrollToNextChunk(newOffset)
		} else {
			break
		}
	}

	if newOffset > 0 {
		beforeRows := w.viewableRows(w.chunk, w.chunkOffset)
		afterRows := w.viewableRows(w.chunk, newOffset)
		if afterRows < w.rows && afterRows < beforeRows {
			cur := w.chunk
			for cur.Next() != nil {
				cur = cur.Next()
			}

			if cur.Type() != logstream.LOADING_CHUNK {
				w.pushRequest(cur.NextRequest(), cur, 1)
			}

			for newOffset > 0 && afterRows < w.rows && afterRows < beforeRows {
				newOffset -= 1
				afterRows = w.viewableRows(w.chunk, newOffset)
			}
		}
	}

	if newOffset == 0 {
		if offset > 0 {
			cur := w.chunk
			for cur.Next() != nil {
				cur = cur.Next()
			}

			if cur.Type() != logstream.LOADING_CHUNK {
				w.pushRequest(cur.NextRequest(), cur, 1)
			}
		}
	}

	w.chunkOffset = newOffset
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

	result := offset - w.chunk.LineCount()

	beforeRows := w.viewableRows(w.chunk, w.chunkOffset)
	afterRows := w.viewableRows(w.chunk.Next(), 0)
	if afterRows < w.rows && afterRows < beforeRows {
		if w.chunk.Viewable() && w.chunkOffset < w.chunk.LineCount() {
			return w.chunk.LineCount() - 1, true
		}
		return 0, false
	}

	w.chunk = w.chunk.Next()
	w.chunkOffset = 0
	return result, true
}

func (w *LogViewPort) pushRequest(request *logstream.ChunkRequest, chunk *logstream.LogChunk, tryToScroll int) {
	if request == nil {
		return
	}

	w.tryToScrollChunk = chunk
	w.tryToScroll = tryToScroll
	w.requests <- *request
}
