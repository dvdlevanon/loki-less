package logstream

func newLogOrigin(labels map[string]string) *LogOrigin {
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
