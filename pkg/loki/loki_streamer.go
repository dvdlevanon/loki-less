package loki

import (
	"fmt"
	"time"

	"github.com/dvdlevanon/loki-less/pkg/logstream"
)

func NewLokiStreamer(stream *logstream.LogStream, requests <-chan logstream.ChunkRequest) *LokiStreamer {
	return &LokiStreamer{
		stream:   stream,
		requests: requests,
	}
}

type LokiStreamer struct {
	stream   *logstream.LogStream
	requests <-chan logstream.ChunkRequest

	chunksCounter int
}

func (s *LokiStreamer) Stream() {
	ticker := time.NewTicker(1 * time.Second)
	labels := make(map[string]string, 0)
	labels["app"] = "test-app"
	labels["pod"] = "test-pod"

	for {
		select {
		case <-ticker.C:
			s.downloadChunk(logstream.ChunkRequest{
				Origin:    s.stream.GetOrCreateOrigin(labels),
				StartNano: time.Now().UnixNano() - 1000000000,
				EndNano:   time.Now().UnixNano(),
				Limit:     10,
			})
		}
	}
	// for {
	// 	select {
	// 	case request := <-s.requests:
	// 		s.downloadChunk(request)
	// 		// s.stream.AddChunk(&chunk)

	// 		// time.Sleep(time.Second / 100)
	// 		// counter++
	// 	}
	// }
}

func (s *LokiStreamer) downloadChunk(req logstream.ChunkRequest) {
	chunk := s.generateChunk(req.Origin.Labels()["app"], req.Origin.Labels()["pod"], req.StartNano, req.EndNano, 10)
	s.stream.AddChunk(&chunk)
}

func (s *LokiStreamer) generateChunk(app string, pod string, startNano int64, endNano int64, linesCount int) logstream.LogChunk {
	labels := make(map[string]string, 0)
	labels["app"] = app
	labels["pod"] = pod
	origin := s.stream.GetOrCreateOrigin(labels)

	lines := make([]logstream.LogLine, 0)
	timeInterval := (endNano - startNano) / int64(linesCount)
	for i := 0; i < linesCount; i++ {
		line := fmt.Sprintf("[%d-%d]\tSome text[app: %s] [pod: %s]", s.chunksCounter, i, app, pod)
		lines = append(lines, logstream.NewLogLine(origin, startNano+(int64(i)*timeInterval), line))
	}

	chunk := s.stream.NewChunk(logstream.RAM_CHUNK)
	chunk.SetLines(lines)
	s.chunksCounter++

	return chunk
}
