package logstream

import (
	"fmt"

	"github.com/dvdlevanon/loki-less/pkg/utils"
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
	return fmt.Sprintf("%s\t%s", utils.FormatNanoTime(l.nanoTime), l.text)
}
