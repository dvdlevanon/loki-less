package logstream

import (
	"sync"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("logstream")

func NewLogStream(origin LogOrigin) *LogStream {
	stream := &LogStream{
		lock:    sync.RWMutex{},
		origins: newOriginList(),
		origin:  origin,
	}

	return stream
}

type LogStream struct {
	head    *LogChunk
	tail    *LogChunk
	lock    sync.RWMutex
	origins originList
	origin  LogOrigin
}

func (ls *LogStream) Origin() *LogOrigin {
	return &ls.origin
}

func (ls *LogStream) Head() *LogChunk {
	return ls.head
}

func (ls *LogStream) Tail() *LogChunk {
	return ls.tail
}

func (ls *LogStream) TimeRequest(nanoTime int64) *ChunkRequest {
	return &ChunkRequest{
		Origin:   ls.Origin(),
		TimeNano: nanoTime,
		Forward:  true,
		Limit:    0,
	}
}

func (ls *LogStream) StartLoading(req ChunkRequest) *LogChunk {
	loadingChunk := req.NewLoadingChunk()
	loadingChunk.StartLoading()
	ls.AddChronological(&loadingChunk)
	return &loadingChunk
}

func (ls *LogStream) GetLoading(req ChunkRequest) *LogChunk {
	cur := ls.head
	for cur != nil {
		if req.IsLoadingChunk(cur) {
			return cur
		}

		cur = cur.next
	}

	return nil
}

func (ls *LogStream) FinishLoading(req ChunkRequest, loaded *LogChunk) *LogChunk {
	loadingChunk := ls.GetLoading(req)
	if loadingChunk == nil {
		return nil
	}

	ls.DeleteChunk(loadingChunk)
	addedChunk := ls.AddChronological(loaded)
	loadingChunk.FinishLoading(addedChunk)
	return addedChunk
}

func (ls *LogStream) GetOrCreateOrigin(labels map[string]string) *LogOrigin {
	return ls.origins.getOrCreate(labels)
}

func (ls *LogStream) AddChronological(chunk *LogChunk) *LogChunk {
	ls.lock.Lock()
	defer ls.lock.Unlock()

	logger.Infof("Add chronological chunk %s", chunk)

	cur := ls.head
	for cur != nil {
		if chunk.IsAfter(cur) {
			cur = cur.Next()
			continue
		}

		break
	}

	if cur == nil {
		ls.appendChunk(chunk)
		return chunk
	}

	if chunk.IsBefore(cur) {
		ls.insertBefore(cur, chunk)
		return chunk
	}

	cur.mergeLines(chunk.lines)
	return cur
}

func (ls *LogStream) appendChunk(chunk *LogChunk) {
	logger.Infof("Appending chunk %s", chunk)

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

func (ls *LogStream) insertBefore(chunk *LogChunk, newChunk *LogChunk) {
	logger.Infof("Inserting chunk %s before %s", newChunk, chunk)

	if chunk == nil || newChunk == nil {
		return
	}

	newChunk.stream = ls
	newChunk.next = chunk
	newChunk.prev = chunk.prev

	if chunk.prev != nil {
		chunk.prev.next = newChunk
	} else {
		ls.head = newChunk
	}
	chunk.prev = newChunk
}

func (ls *LogStream) DeleteChunk(chunk *LogChunk) {
	logger.Infof("Deleting chunk %s", chunk)

	ls.lock.Lock()
	defer ls.lock.Unlock()

	if chunk == nil {
		return
	}

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

	chunk.next = nil
	chunk.prev = nil
	chunk.stream = nil
}
