package loki

import (
	"fmt"
	"time"

	"github.com/dvdlevanon/loki-less/pkg/logstream"
)

func NewLokiStreamer(stream *logstream.LogStream) *LokiStreamer {
	return &LokiStreamer{
		stream: stream,
	}
}

type LokiStreamer struct {
	stream *logstream.LogStream
}

func (s *LokiStreamer) Stream() {
	counter := 0
	for {
		labels := make(map[string]string, 0)
		labels["app"] = "api-service"
		labels["env"] = "staging"
		origin := s.stream.GetOrCreateOrigin(labels)

		lines := make([]logstream.LogLine, 0)
		for i := 0; i < 10; i++ {
			line := fmt.Sprintf("%d Some line %d", counter, i)
			lines = append(lines, logstream.NewLogLine(origin, time.Now().UnixNano(), line))
		}

		chunk := s.stream.NewChunk(logstream.RAM_CHUNK)
		chunk.SetLines(lines)

		s.stream.AddChunk(&chunk)

		time.Sleep(time.Second / 100)
		counter++
	}
}
