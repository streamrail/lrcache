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
	"flag"
	"fmt"
	"strconv"
	"testing"
)

var (
	redisHostBad      = flag.String("redis_bad", "loc22alhost:6379", "Redis host and port. Eg: localhost:6379")
	redisHost         = flag.String("redis", "localhost:6379", "Redis host and port. Eg: localhost:6379")
	redisConnPoolSize = flag.Int("redisConnPoolSize", 5, "Redis connection pool size. Default: 5")
	redisPrefix       = flag.String("redisPrefix", "rl_", "Redis prefix to attach to keys")
	lruSize           = flag.Int("lruSize", 5, "LRU Cache Size. Default: 5")
)

func TestLRCacheEvict(t *testing.T) {
	cache, _ := NewLRCache(*redisHost, *redisConnPoolSize, *redisPrefix, *lruSize)

	// set 10 elements in the cache. since the LRU size is set to 5, we expect 5 evictions from LRU
	evictions := 0
	for i := 0; i < 10; i++ {
		key := strconv.Itoa(i)
		value := i
		if cache.Set(key, value) {
			evictions++
		}
	}
	if evictions != 5 {
		t.Fatalf("should have evicted 5 items from LRU")
	}
}

// set 10 elements in the cache. since the LRU size is set to 5,
// we expect 5 elements in Redis for the first items added (coldest ones)
func TestLRCacheKeyInRedis(t *testing.T) {
	cache, _ := NewLRCache(*redisHost, *redisConnPoolSize, *redisPrefix, *lruSize)

	for i := 0; i < 10; i++ {
		key := strconv.Itoa(i)
		value := i
		cache.Set(key, value)
	}
	for i := 0; i < 5; i++ {
		result := cache.Get(strconv.Itoa(i))
		val, err := result.IntVal()
		if err == nil {
			if !result.FromRedis() {
				t.Fatalf("value came from LRU: %d, but it was expected to get it from Redis. Are you running a redis server?\n", val)
			}
		} else {
			fmt.Println(err.Error())
		}
	}
}

// set 10 elements in the cache. since the LRU size is set to 5,
// we expect 5 elements in LRU for the last items added (hottest ones)
func TestLRCacheKeyInLRU(t *testing.T) {
	cache, _ := NewLRCache(*redisHost, *redisConnPoolSize, *redisPrefix, *lruSize)

	for i := 0; i < 10; i++ {
		key := strconv.Itoa(i)
		value := i
		cache.Set(key, value)
	}
	for i := 5; i < 10; i++ {
		result := cache.Get(strconv.Itoa(i))
		_, err := result.IntVal()
		if err == nil {
			if result.FromRedis() {
				t.Fatalf("value came from Redis but it was expected to get it from LRU")
			}
		} else {
			t.Fatalf(err.Error())
		}
	}
}

// set 10 elements in the cache. since there is no connection to redis, all should come from LRU
func TestLRCacheNoRedisHost(t *testing.T) {
	cache, _ := NewLRCache(*redisHostBad, *redisConnPoolSize, *redisPrefix, *lruSize)

	for i := 0; i < 10; i++ {
		key := strconv.Itoa(i)
		value := i
		cache.Set(key, value)
	}
	for i := 0; i < 10; i++ {
		result := cache.Get(strconv.Itoa(i))
		_, err := result.IntVal()
		if err == nil {
			if result.FromRedis() {
				t.Fatalf("value came from Redis but it was expected to get it from LRU")
			}
		} else {
			t.Fatalf(err.Error())
		}
	}
}
