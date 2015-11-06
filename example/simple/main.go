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

package main

import (
	"flag"
	"fmt"
	"github.com/streamrail/lrcache"
	"strconv"
)

var (
	redisHost         = flag.String("redis", "localhost:6379", "Redis host and port. Eg: localhost:6379")
	redisConnPoolSize = flag.Int("redisConnPoolSize", 5, "Redis connection pool size. Default: 5")
	redisPrefix       = flag.String("redisPrefix", "rl_", "Redis prefix to attach to keys")
	lruSize           = flag.Int("lruSize", 5, "LRU Cache Size. Default: 5")
)

func main() {
	cache, _ := lrcache.NewLRCache(*redisHost, *redisConnPoolSize, *redisPrefix, *lruSize)

	// set 10 elements in the cache. since the LRU size is set to 5,
	// we expect 5 elements in the lru, 5 evictions from lru, and 5 elements in redis
	for i := 0; i < 10; i++ {
		key := strconv.Itoa(i)
		value := i
		fmt.Printf("setting key %s with value %d. evicted stuff from cache? %v\n", key, value, cache.Set(key, value))
	}
	for i := 0; i < 10; i++ {
		result := cache.Get(strconv.Itoa(i))
		val, err := result.IntVal()
		if err == nil {
			if result.FromRedis() {
				fmt.Printf("value came from redis: %d\n", val)
			} else {
				fmt.Printf("value came from LRU: %d\n", val)
			}
		} else {
			fmt.Println(err.Error())
		}
	}
}
