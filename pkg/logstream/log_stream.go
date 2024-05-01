package logstream

import (
	"sync"
)

func NewLogStream() *LogStream {
	stream := &LogStream{
		lock:    sync.RWMutex{},
		origins: newOriginList(),
	}

	// chnuk := stream.NewChunk(HOLE_CHUNK)
	// stream.AddChunk(&chnuk)

	return stream
}

type LogStream struct {
	head    *LogChunk
	tail    *LogChunk
	lock    sync.RWMutex
	origins originList
}

func (ls *LogStream) Head() *LogChunk {
	return ls.head
}

func (ls *LogStream) Tail() *LogChunk {
	return ls.tail
}

func (ls *LogStream) NewChunk(t ChunkType) LogChunk {
	return newLogChunk(t)
}

func (ls *LogStream) AddChunk(chunk *LogChunk) {
	ls.lock.Lock()
	defer ls.lock.Unlock()

	chunk.stream = ls
	if ls.tail == nil {
		ls.head = chunk
		ls.tail = chunk
	} else {
		chunk.prev = ls.tail
		ls.tail.next = chunk
		ls.tail = chunk
	}
}

func (ls *LogStream) InsertChunk(after *LogChunk, chunk *LogChunk) {
	ls.lock.Lock()
	defer ls.lock.Unlock()

	if after == nil {
		return
	}

	chunk.stream = ls
	chunk.prev = after
	chunk.next = after.next

	if after.next != nil {
		after.next.prev = chunk
	}
	after.next = chunk

	if ls.tail == after {
		ls.tail = chunk
	}
}

func (ls *LogStream) RemoveChunk(chunk *LogChunk) {
	ls.lock.Lock()
	defer ls.lock.Unlock()

	if chunk.prev != nil {
		chunk.prev.next = chunk.next
	} else {
		ls.head = chunk.next
	}

	if chunk.next != nil {
		chunk.next.prev = chunk.prev
	} else {
		ls.tail = chunk.prev
	}

	chunk.stream = nil
}

func (ls *LogStream) ReplaceChunk(oldChunk, newChunk *LogChunk) {
	ls.lock.Lock()
	defer ls.lock.Unlock()

	newChunk.prev = oldChunk.prev
	newChunk.next = oldChunk.next

	if oldChunk.prev != nil {
		oldChunk.prev.next = newChunk
	} else {
		ls.head = newChunk
	}

	if oldChunk.next != nil {
		oldChunk.next.prev = newChunk
	} else {
		ls.tail = newChunk
	}

	oldChunk.stream = nil
	newChunk.stream = ls
}

func (ls *LogStream) GetOrCreateOrigin(labels map[string]string) *LogOrigin {
	return ls.origins.getOrCreate(labels)
}
