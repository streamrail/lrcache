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
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/streamrail/lrcache"
	"log"
)

var (
	redisHost         = flag.String("redis", "localhost:6379", "Redis host and port. Eg: localhost:6379")
	redisConnPoolSize = flag.Int("redisConnPoolSize", 5, "Redis connection pool size. Default: 5")
	redisPrefix       = flag.String("redisPrefix", "rl_", "Redis prefix to attach to keys")

	// the LRU size is set to 1
	lruSize = flag.Int("lruSize", 1, "LRU Cache Size. Default: 1")
)

type Person struct {
	Name string
	Age  int
}

func main() {
	cache, _ := lrcache.NewLRCache(*redisHost, *redisConnPoolSize, *redisPrefix, *lruSize)

	bob := &Person{Name: "Bob", Age: 21}
	sally := &Person{Name: "Alice", Age: 19}

	cache.Set("bob-key", bob)     // pushes bob-key to LRU
	cache.Set("alice-key", sally) // pushes bob-key out of LRU and into Redis, puts alice-key in LRU

	res1 := cache.Get("bob-key")
	res2 := cache.Get("alice-key")

	log.Println(ExtractCustomType(res1)) // from redis
	log.Println(ExtractCustomType(res2)) // from LRU

}

func ExtractCustomType(res *lrcache.LRCacheItem) (*Person, error) {
	if res.Error() != nil {
		return nil, res.Error()
	}
	if res.Value() == nil {
		return nil, nil
	}
	if !res.FromRedis() { // value from LRU Cache
		log.Printf("from LRU cache\n")
		if v, ok := res.Value().(*Person); ok {
			return v, nil
		} else {
			return nil, fmt.Errorf("result does not contain value *Person")
		}
	} else {
		log.Printf("from Redis cache\n")
		var decoded = new(Person)
		if gobVal, ok := res.Value().([]uint8); ok {
			dec := gob.NewDecoder(bytes.NewBuffer(gobVal))
			err := dec.Decode(decoded)
			if err != nil {
				return nil, err
			} else {
				return decoded, nil
			}
		} else {
			return nil, fmt.Errorf("expected result to contain gob but got other type")
		}
	}
}
