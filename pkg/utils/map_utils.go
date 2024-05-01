package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"sort"
)

func HashMap(data map[string]string) string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	hasher := md5.New()

	for _, key := range keys {
		io.WriteString(hasher, key+"="+data[key]+";")
	}

	return fmt.Sprintf("%x", hasher.Sum(nil))
}
