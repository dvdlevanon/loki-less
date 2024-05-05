package streamer

import "github.com/dvdlevanon/loki-less/pkg/logstream"

type SourceResponse struct {
	Request logstream.ChunkRequest
	Chunk   logstream.LogChunk
}

type Source interface {
	Next(req logstream.ChunkRequest, doneChannel chan<- SourceResponse)
}
