package logstream

import "fmt"

func NewLogOrigin(labels map[string]string) *LogOrigin {
	return &LogOrigin{
		labels: labels,
	}
}

type LogOrigin struct {
	labels map[string]string
}

func (o *LogOrigin) Labels() map[string]string {
	return o.labels
}

func (o *LogOrigin) String() string {
	return fmt.Sprintf("Origin[%s]", o.labels)
}
