package logstream

type ChunkRequest struct {
	Origin    *LogOrigin
	StartNano int64
	EndNano   int64
	Limit     int
}
