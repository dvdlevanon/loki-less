package utils

import "time"

func FormatNanoTime(nanoTime int64) string {
	return time.Unix(0, nanoTime).Format("2006-01-02 15:04:05")
}
