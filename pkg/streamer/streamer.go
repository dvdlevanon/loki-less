package streamer

import (
	"github.com/dvdlevanon/loki-less/pkg/logstream"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("streamer")

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

	logger.Infof("Requesting new chunk %s", req.String())
	s.stream.StartLoading(req)
	go s.source.Next(req, s.responses)
}

func (s *Streamer) handleResponse(resp *SourceResponse) {
	logger.Infof("New chunk response [req: %s] [loaded: %s]", &resp.Request, &resp.Chunk)
	s.stream.FinishLoading(resp.Request, &resp.Chunk)
}
