package redshift

import (
	"sync"
	"time"
)

const defaultLongevity time.Duration = time.Second * 10

type DataCache struct {
	mapping map[string][]byte
	lock    sync.RWMutex
}

func (this *DataCache) Add(key string, b []byte, longevity time.Duration) {
	if longevity == 0 {
		longevity = defaultLongevity

	}
	this.lock.Lock()
	defer this.lock.Unlock()
	this.mapping[key] = b

	go func() {
		time.Sleep(longevity)
		this.lock.Lock()
		defer this.lock.Unlock()
		delete(this.mapping, key)
	}()
}

func (this *DataCache) Get(key string) (b []byte, ok bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	b, ok = this.mapping[key] //one-liner won't work for some reason o_O
	return
}

func NewDataCache() *DataCache {
	return &DataCache{mapping: make(map[string][]byte)}
}
