package logstream

import (
	"fmt"
	"time"
)

func NewLogLine(origin *LogOrigin, nano_time int64, line string) LogLine {
	return LogLine{
		origin:   origin,
		nanoTime: nano_time,
		text:     line,
	}
}

type LogLine struct {
	origin   *LogOrigin
	nanoTime int64
	text     string
}

func (l *LogLine) Origin() *LogOrigin {
	return l.origin
}

func (l *LogLine) NanoTime() int64 {
	return l.nanoTime
}

func (l *LogLine) Text() string {
	return l.text
}

func (l *LogLine) FormattedLine() string {
	t := time.Unix(0, l.nanoTime)
	formattedTime := t.Format("2006-01-02 15:04:05")

	return fmt.Sprintf("%s\t%s", formattedTime, l.text)
}
