package main

import (
	"cacheSys/cache"
	"time"
)

func main() {
	cache := cache.NewMemCache()
	cache.SetMaxMemory("100MB")

	cache.Set("int", 1, time.Second)
	cache.Set("bool", false, time.Second)
	cache.Set("data", map[string]interface{}{"a": 1}, time.Second)

	cache.Get("int")
	cache.Del("int")
	cache.Flush()
	cache.Keys()
}
