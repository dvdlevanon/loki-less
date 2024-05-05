package streamer

import (
	"github.com/dvdlevanon/loki-less/pkg/logstream"
)

func NewStreamer(stream *logstream.LogStream, requests <-chan logstream.ChunkRequest, source Source) *Streamer {
	return &Streamer{
		stream:    stream,
		requests:  requests,
		source:    source,
		responses: make(chan SourceResponse),
	}
}

type Streamer struct {
	stream    *logstream.LogStream
	requests  <-chan logstream.ChunkRequest
	source    Source
	responses chan SourceResponse
}

func (s *Streamer) Stream() {
	// ticker := time.NewTicker(1 * time.Second)
	// labels := make(map[string]string, 0)
	// labels["app"] = "test-app"
	// labels["pod"] = "test-pod"

	// for {
	// 	select {
	// 	case <-ticker.C:
	// 		s.downloadChunk(logstream.ChunkRequest{
	// 			Origin:   s.stream.GetOrCreateOrigin(labels),
	// 			TimeNano: time.Now().UnixNano(),
	// 			Forward:  true,
	// 			Limit:    10,
	// 		})
	// 	}
	// }

	for {
		select {
		case request := <-s.requests:
			s.handleRequest(request)
		case response := <-s.responses:
			s.handleResponse(&response)
		}
	}
}

func (s *Streamer) handleRequest(req logstream.ChunkRequest) {
	loadingChunk := s.stream.GetLoading(req)
	if loadingChunk != nil {
		return
	}

	s.stream.StartLoading(req)
	go s.source.Next(req, s.responses)
}

func (s *Streamer) handleResponse(resp *SourceResponse) {
	s.stream.FinishLoading(resp.Request, &resp.Chunk)
}
