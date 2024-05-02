package ui

import "github.com/dvdlevanon/loki-less/pkg/logstream"

func NewLogViewPort(stream *logstream.LogStream) *LogViewPort {
	return &LogViewPort{
		stream: stream,
	}
}

type LogViewPort struct {
	stream      *logstream.LogStream
	rows        int
	chunk       *logstream.LogChunk
	chunkOffset int
	autoScroll  bool
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
		// request current
		return nil, 0
	}

	if v.autoScroll {
		v.ScrollToLatest()
	}

	return v.chunk, v.chunkOffset
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
	v.chunkOffset = v.chunk.LineCount()
	shownLines := 0

	for shownLines <= v.rows {
		v.chunkOffset = v.chunkOffset - 1
		shownLines = shownLines + 1
		if v.chunkOffset < 0 {
			v.chunk = v.chunk.Prev()
			v.chunkOffset = v.chunk.LineCount()
		}
	}
}

func (w *LogViewPort) Scroll(offset int) {
	newOffset := w.chunkOffset + offset

	for {
		if newOffset < 0 {
			if w.chunk.Prev() == nil {
				// request chunk
				return
			}

			w.chunk = w.chunk.Prev()

			if !w.chunk.Viewable() {
				return
			}

			newOffset = w.chunk.LineCount() + newOffset
		} else if !w.chunk.Viewable() {
			return
		} else if newOffset > w.chunk.LineCount() {
			if w.chunk.Next() == nil {
				// requset next
				return
			}

			newOffset = newOffset - w.chunk.LineCount()
			w.chunk = w.chunk.Next()
		} else {
			break
		}
	}

	w.chunkOffset = newOffset
}
