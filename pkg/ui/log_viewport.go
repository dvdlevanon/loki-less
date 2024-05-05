package ui

import (
	"github.com/dvdlevanon/loki-less/pkg/logstream"
)

func NewLogViewPort(stream *logstream.LogStream, requests chan<- logstream.ChunkRequest) *LogViewPort {
	return &LogViewPort{
		stream:   stream,
		requests: requests,
	}
}

type LogViewPort struct {
	stream      *logstream.LogStream
	rows        int
	chunk       *logstream.LogChunk
	chunkOffset int
	autoScroll  bool
	requests    chan<- logstream.ChunkRequest
	tryToScroll int
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
		// v.pushRequest(v.stream.TimeRequest(time.Now().UnixNano()))
		v.pushRequest(v.stream.TimeRequest(1000000000000000000), 0)
		return nil, 0
	}

	v.refreshLoadingChunk()
	v.refreshTryToScroll()

	if v.autoScroll {
		v.ScrollToLatest()
		v.pushRequest(v.stream.Tail().NextRequest(), 1)
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
	}
}

func (v *LogViewPort) refreshTryToScroll() {
	if v.tryToScroll == 0 {
		return
	}

	chunk := v.chunk
	if v.tryToScroll < 0 {
		chunk = chunk.Prev()
	} else {
		chunk = chunk.Next()
	}

	if chunk == nil {
		return
	}

	v.chunk = chunk
	v.chunkOffset = 0
	if v.tryToScroll < 0 {
		v.chunkOffset = v.chunk.LineCount()
	}

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

	for shownLines <= v.rows {
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
	newOffset := w.chunkOffset + offset

	chunk := w.chunk
	for chunk != nil && !chunk.Viewable() {
		if newOffset > 0 {
			chunk = chunk.Next()
		} else {
			chunk = chunk.Prev()
		}
	}

	if chunk != nil {
		w.chunk = chunk
	}

	for {
		if w.chunk == nil || !w.chunk.Viewable() {
			return
		} else if newOffset < 0 {
			newOffset = w.scrollToPrevChunk(newOffset)
		} else if newOffset > w.chunk.LineCount() {
			newOffset = w.scrollToNextChunk(newOffset)
		} else {
			break
		}
	}

	w.chunkOffset = newOffset
}

func (w *LogViewPort) scrollToPrevChunk(offset int) int {
	if w.chunk.Prev() == nil {
		w.pushRequest(w.chunk.PrevRequest(), -1)
		return 0
	}

	w.chunk = w.chunk.Prev()

	if !w.chunk.Viewable() {
		return 0
	}

	return w.chunk.LineCount() + offset
}

func (w *LogViewPort) scrollToNextChunk(offset int) int {
	if w.chunk.Next() == nil {
		w.pushRequest(w.chunk.NextRequest(), 1)
		return 0
	}

	result := offset - w.chunk.LineCount()
	w.chunk = w.chunk.Next()
	return result
}

func (w *LogViewPort) pushRequest(request *logstream.ChunkRequest, tryToScroll int) {
	if request == nil {
		return
	}

	w.tryToScroll = tryToScroll
	w.requests <- *request
}
