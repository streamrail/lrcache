// Copyright (c) 2015 streamrail

// The MIT License (MIT)

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package lrcache

import (
	"github.com/hashicorp/golang-lru"
	redis "github.com/streamrail/redis-storage"
)

// LRCache holds a reference to an LRU cache and a Redis cache
type LRCache struct {
	lruCache   *lru.Cache
	redisCache *redis.RedisStorage
	onEvict    func(key interface{}, val interface{})
}

// Gets a new instance of LRCache, providing all the reqired params for Redis and LRU Cache
func NewLRCache(redisHost string, redisConnPoolSize int, redisPrefix string, size int) (*LRCache, error) {
	result := &LRCache{
		lruCache:   nil,
		redisCache: redis.NewRedisStorage(redisHost, redisConnPoolSize, redisPrefix),
	}
	// When evictning an item from the in-memoro LRU cache, save the item on redis as a second caching tier
	result.onEvict = func(key interface{}, val interface{}) {
		if str, ok := key.(string); ok {
			result.redisCache.Set(str, val)
		}
	}
	l, err := lru.NewWithEvict(size, result.onEvict)
	result.lruCache = l
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Gets an item from lru cache if it's there. if it's missing from lru cache,
// try go get it from redis and then set it in the lru cache for next time
func (lr *LRCache) Get(key string) *LRCacheItem {
	if lr.lruCache.Contains(key) { // found on lru
		if v, ok := lr.lruCache.Get(key); ok {
			return newLRCacheItem(v, nil, false)
		}
		// LRU doesn't have value anymore, try redis
		v, err := lr.redisCache.Get(key)
		if err == nil { // found on redis
			return newLRCacheItem(v, nil, true)
		}
		return newLRCacheItem(nil, err, false)
	} else { // not in lru, try redis
		v, err := lr.redisCache.Get(key)
		if err == nil { // found on redis
			return newLRCacheItem(v, nil, true)
		}
		return newLRCacheItem(nil, err, false)
	}
}

// Set in cache. Returns true if key was evicted from LRU and put in Redis
func (lr *LRCache) Set(key string, val interface{}) bool {
	return lr.lruCache.Add(key, val)
}

func (lr *LRCache) Delete(key string) {
	lr.lruCache.Remove(key)
	lr.redisCache.Delete(key)
}
