package source

import (
	"fmt"
	"time"

	"github.com/dvdlevanon/loki-less/pkg/logstream"
	"github.com/dvdlevanon/loki-less/pkg/streamer"
)

func NewFakeSource(delay time.Duration) *FakeSource {
	return &FakeSource{delay: delay}
}

type FakeSource struct {
	delay         time.Duration
	chunksCounter int
}

func (s *FakeSource) Next(req logstream.ChunkRequest, doneChannel chan<- streamer.SourceResponse) {
	time.Sleep(s.delay)
	lines := make([]logstream.LogLine, 0)

	timeInterval := 1000000000

	for i := 0; i < req.Limit; i++ {
		line := fmt.Sprintf("[%d-%d]\tSome text from origin: %s", s.chunksCounter, i, req.Origin)
		if req.Forward {
			lines = append(lines, logstream.NewLogLine(req.Origin, req.TimeNano+(int64(i*timeInterval)), line))
		} else {
			lines = append(lines, logstream.NewLogLine(req.Origin, req.TimeNano-(int64(i*timeInterval)), line))
		}
	}

	chunk := logstream.NewRamLogChunk(lines)
	s.chunksCounter++

	doneChannel <- streamer.SourceResponse{
		Request: req,
		Chunk:   chunk,
	}
}
