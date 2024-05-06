package source

import (
	"fmt"
	"time"

	"github.com/dvdlevanon/loki-less/pkg/logstream"
	"github.com/dvdlevanon/loki-less/pkg/streamer"
)

func NewFakeSource(delay time.Duration, defaultLimit, intersectFactor int) *FakeSource {
	return &FakeSource{
		delay:           delay,
		defaultLimit:    defaultLimit,
		intersectFactor: intersectFactor,
	}
}

type FakeSource struct {
	delay           time.Duration
	defaultLimit    int
	chunksCounter   int
	intersectFactor int
}

func (s *FakeSource) getLinesCount(req logstream.ChunkRequest) int {
	if req.Limit == 0 {
		return s.defaultLimit
	} else {
		return req.Limit
	}
}

func (s *FakeSource) getNanoTime(req logstream.ChunkRequest, interval, linesCount int) int64 {
	if s.intersectFactor == 0 {
		return req.TimeNano
	}

	return req.TimeNano + (req.Direction()*-1)*int64((linesCount/s.intersectFactor)*interval)
}

func (s *FakeSource) Next(req logstream.ChunkRequest, doneChannel chan<- streamer.SourceResponse) {
	time.Sleep(s.delay)
	lines := make([]logstream.LogLine, 0)

	timeInterval := 1000000000
	linesCount := s.getLinesCount(req)
	nanoTime := s.getNanoTime(req, timeInterval, linesCount)

	for i := 0; i < linesCount; i++ {
		line := fmt.Sprintf("[%d-%d]\tSome text from origin: %s", s.chunksCounter, i, req.Origin)
		lines = append(lines, logstream.NewLogLine(req.Origin, nanoTime+(req.Direction()*(int64(i*timeInterval))), line))
	}

	chunk := logstream.NewRamLogChunk(lines)
	s.chunksCounter++

	if chunk == nil {
		return
	}

	doneChannel <- streamer.SourceResponse{
		Request: req,
		Chunk:   *chunk,
	}
}
