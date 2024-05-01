package logstream

type ChunkType int

const (
	RAM_CHUNK ChunkType = iota + 1
	DISK_CHUNK
	HOLE_CHUNK
	ERROR_CHUNK
	LOADING_CHUNK
)

func newLogChunk(t ChunkType) LogChunk {
	return LogChunk{chunkType: t}
}

type LogChunk struct {
	stream    *LogStream
	chunkType ChunkType
	lines     []LogLine
	next      *LogChunk
	prev      *LogChunk
}

func (c *LogChunk) Stream() *LogStream {
	return c.stream
}

func (c *LogChunk) Prev() *LogChunk {
	return c.prev
}

func (c *LogChunk) Next() *LogChunk {
	return c.next
}

func (c *LogChunk) Type() ChunkType {
	return c.chunkType
}

func (c *LogChunk) Lines() []LogLine {
	return c.lines
}

func (c *LogChunk) LineCount() int {
	return len(c.lines)
}

func (c *LogChunk) SetLines(lines []LogLine) {
	// sort
	c.lines = lines
}
