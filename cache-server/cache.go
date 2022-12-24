package cache_server

import (
	"cacheSys/cache"
	"time"
)

type cacheServer struct {
	cacheServer cache.Cache
}

func NewMemCache() *cacheServer {
	return &cacheServer{
		cacheServer: cache.NewMemCache(),
	}
}

func (cs *cacheServer) SetMaxMemory(size string) bool {
	return cs.memCache.SetMaxMemory(size)
}

func (cs *cacheServer) Set(key string, val interface{}, expire ...time.Duration) bool {
	expireTs := time.Second * 0
	if len(expire) > 0 {
		expireTs = expire[0]
	}
	return cs.memCache.Set(key, val, expireTs)
}

func (cs *cacheServer) get(key string) (*cacheServerValue, bool) {
	val, ok := cs.values[key]
	return val, ok
}

func (cs *cacheServer) del(key string) {
	tmp, ok := cs.get(key)
	if ok && tmp != nil {
		cs.currMemorySize -= tmp.size
		delete(cs.values, key)
	}
}

func (cs *cacheServer) add(key string, val *cacheServerValue) {
	cs.values[key] = val
	cs.currMemorySize += val.size
}

func (cs *cacheServer) Get(key string) (interface{}, bool) {
	cs.locker.RLock()
	defer cs.locker.RUnlock()
	csv, ok := cs.get(key)
	if ok {
		if csv.expire != 0 && csv.expireTime.Before(time.Now()) {
			cs.del(key)
			return nil, false
		}
		return csv.val, ok
	}
	return nil, false
}

func (cs *cacheServer) Del(key string) bool {
	cs.locker.Lock()
	defer cs.locker.Unlock()
	cs.del(key)
	return false
}

func (cs *cacheServer) Exists(key string) bool {
	cs.locker.RLock()
	defer cs.locker.RUnlock()
	_, ok := cs.values[key]
	return ok
}

func (cs *cacheServer) Flush() bool {
	cs.locker.Lock()
	defer cs.locker.Unlock()

	cs.values = make(map[string]*cacheServerValue, 0)
	cs.currMemorySize = 0

	return true
}

func (cs *cacheServer) Keys() int64 {
	cs.locker.RLock()
	defer cs.locker.RUnlock()
	return int64(len(cs.values))
}

func (cs *cacheServer) clearExpiredItem() {
	timeTicker := time.NewTicker(cs.clearExpiredItemTimeInterval)
	defer timeTicker.Stop()

	for {
		select {
		case <-timeTicker.C:
			for key, item := range cs.values {
				if item.expire != 0 && time.Now().After(item.expireTime) {
					cs.locker.Lock()
					cs.del(key)
					cs.locker.Unlock()
				}
			}
		}
	}
}
