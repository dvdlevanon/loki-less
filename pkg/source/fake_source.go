package source

import (
	"fmt"
	"time"

	"github.com/dvdlevanon/loki-less/pkg/logstream"
	"github.com/dvdlevanon/loki-less/pkg/streamer"
)

func NewFakeSource(delay time.Duration, defaultLimit int) *FakeSource {
	return &FakeSource{delay: delay, defaultLimit: defaultLimit}
}

type FakeSource struct {
	delay         time.Duration
	defaultLimit  int
	chunksCounter int
}

func (s *FakeSource) Next(req logstream.ChunkRequest, doneChannel chan<- streamer.SourceResponse) {
	time.Sleep(s.delay)
	lines := make([]logstream.LogLine, 0)

	timeInterval := 1000000000
	actualLimit := req.Limit
	if actualLimit == 0 {
		actualLimit = s.defaultLimit
	}

	for i := 0; i < actualLimit; i++ {
		line := fmt.Sprintf("[%d-%d]\tSome text from origin: %s", s.chunksCounter, i, req.Origin)
		if req.Forward {
			lines = append(lines, logstream.NewLogLine(req.Origin, req.TimeNano+(int64(i*timeInterval)), line))
		} else {
			lines = append(lines, logstream.NewLogLine(req.Origin, req.TimeNano-(int64(i*timeInterval)), line))
		}
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
