package logstream

import (
	"sync"

	"github.com/dvdlevanon/loki-less/pkg/utils"
)

func newOriginList() originList {
	return originList{
		origins: make(map[string]*LogOrigin),
		lock:    sync.RWMutex{},
	}
}

type originList struct {
	origins map[string]*LogOrigin
	lock    sync.RWMutex
}

func (o *originList) get(key string) *LogOrigin {
	o.lock.RLock()
	defer o.lock.RUnlock()

	return o.origins[key]
}

func (o *originList) create(key string, labels map[string]string) *LogOrigin {
	newOrigin := NewLogOrigin(labels)

	o.lock.Lock()
	defer o.lock.Unlock()

	o.origins[key] = newOrigin

	return o.origins[key]
}

func (o *originList) getOrCreate(labels map[string]string) *LogOrigin {
	key := utils.HashMap(labels)

	result := o.get(key)
	if result != nil {
		return result
	}

	return o.create(key, labels)
}
