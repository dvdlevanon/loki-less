package logstream

import "fmt"

type ChunkRequest struct {
	Origin   *LogOrigin
	TimeNano int64
	Forward  bool
	Limit    int
}

func (r *ChunkRequest) NewLoadingChunk() LogChunk {
	return NewLoadingChunk(r.TimeNano, r.Forward)
}

func (r *ChunkRequest) IsLoadingChunk(chunk *LogChunk) bool {
	return chunk != nil && chunk.chunkType == LOADING_CHUNK &&
		chunk.nanoTime == r.TimeNano && chunk.loadingForward == r.Forward
}

func (r *ChunkRequest) String() string {
	return fmt.Sprintf("Chunk Request (origin: %s) (time: %d) (forward? %t) (limit: %d)",
		r.Origin, r.TimeNano, r.Forward, r.Limit)
}
