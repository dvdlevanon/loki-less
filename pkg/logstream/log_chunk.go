package logstream

import (
	"fmt"
	"sort"
	"time"
)

type ChunkType int

const (
	RAM_CHUNK ChunkType = iota + 1
	DISK_CHUNK
	LOADING_CHUNK
)

func NewLoadingChunk(nanoTime int64, loadingForward bool) LogChunk {
	return LogChunk{chunkType: LOADING_CHUNK, nanoTime: nanoTime, Abba: fmt.Sprintf("LOADING %d", nanoTime), loadingForward: loadingForward}
}

func NewRamLogChunk(lines []LogLine) *LogChunk {
	if len(lines) == 0 {
		logger.Warning("Refute to create a chunk with zero lines")
		return nil
	}

	sort.Slice(lines, func(i, j int) bool {
		return lines[i].nanoTime < lines[j].nanoTime
	})

	return &LogChunk{chunkType: RAM_CHUNK, lines: lines, Abba: fmt.Sprintf("RAM %d-%d", lines[0].nanoTime, lines[len(lines)-1].nanoTime)}
}

type LogChunk struct {
	Abba      string
	stream    *LogStream
	chunkType ChunkType
	next      *LogChunk
	prev      *LogChunk

	//  ram only
	lines []LogLine

	// disk only
	filename      string
	lowerNanoTime int64
	upperNanoTime int64
	linesCount    int

	// loading only
	nanoTime           int64
	loadingStartMillis int64
	loadedChunk        *LogChunk
	loadingForward     bool
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
	if c.chunkType == RAM_CHUNK {
		return len(c.lines)
	}

	if c.chunkType == DISK_CHUNK {
		return c.linesCount
	}

	if c.chunkType == LOADING_CHUNK {
		return 1
	}

	return 0
}

func (c *LogChunk) IsPrevViewable() bool {
	return c.prev != nil && c.prev.Viewable()
}

func (c *LogChunk) IsNextViewable() bool {
	return c.next != nil && c.next.Viewable()
}

func (c *LogChunk) Viewable() bool {
	return c.chunkType == RAM_CHUNK || c.chunkType == DISK_CHUNK
}

func (c *LogChunk) PrevRequest() *ChunkRequest {
	if c.chunkType != RAM_CHUNK && c.chunkType != DISK_CHUNK {
		return nil
	}

	return &ChunkRequest{
		Origin:   c.stream.Origin(),
		TimeNano: c.lines[0].nanoTime - 1,
		Forward:  false,
		Limit:    0,
	}
}

func (c *LogChunk) NextRequest() *ChunkRequest {
	if c.chunkType != RAM_CHUNK && c.chunkType != DISK_CHUNK {
		return nil
	}

	return &ChunkRequest{
		Origin:   c.stream.Origin(),
		TimeNano: c.lines[len(c.lines)-1].nanoTime + 1,
		Forward:  true,
		Limit:    0,
	}
}

func (c *LogChunk) LowerNanoTime() int64 {
	if c.chunkType == RAM_CHUNK {
		return c.lines[0].nanoTime
	}

	if c.chunkType == DISK_CHUNK {
		return c.lowerNanoTime
	}

	if c.chunkType == LOADING_CHUNK {
		return c.nanoTime
	}

	return 0
}

func (c *LogChunk) UpperNanoTime() int64 {
	if c.chunkType == RAM_CHUNK {
		return c.lines[len(c.lines)-1].nanoTime
	}

	if c.chunkType == DISK_CHUNK {
		return c.upperNanoTime
	}

	if c.chunkType == LOADING_CHUNK {
		return c.nanoTime
	}

	return 0
}

func (c *LogChunk) IsAfter(other *LogChunk) bool {
	return c.UpperNanoTime() > other.LowerNanoTime()
}

func (c *LogChunk) IsBefore(other *LogChunk) bool {
	return c.UpperNanoTime() < other.LowerNanoTime()
}

func (c *LogChunk) LoadedChunk() *LogChunk {
	return c.loadedChunk
}

func (c *LogChunk) StartLoading() {
	c.loadingStartMillis = time.Now().UnixMilli()
}

func (c *LogChunk) FinishLoading(loaded *LogChunk) {
	c.loadingStartMillis = -1
	c.loadedChunk = loaded
}

func (c *LogChunk) ElapsedLoadingTime() int64 {
	return time.Now().UnixMilli() - c.loadingStartMillis
}

func (c *LogChunk) LoadingForward() bool {
	return c.loadingForward
}

func (c *LogChunk) String() string {
	if c.chunkType == LOADING_CHUNK {
		if c.loadedChunk != nil {
			return fmt.Sprintf("Loaded - %s", c.loadedChunk)
		} else {
			return fmt.Sprintf("Loading (%dms) (forward? %t)", c.ElapsedLoadingTime(), c.LoadingForward())
		}
	}

	if c.chunkType == RAM_CHUNK {
		return fmt.Sprintf("Ram %d lines", c.LineCount())
	}

	return "Chunk String Not implemented"
}
