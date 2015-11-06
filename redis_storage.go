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
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/garyburd/redigo/redis"
)

type redisStorage struct {
	pool   *redis.Pool
	prefix string
}

func (r *redisStorage) connection() redis.Conn {
	return r.pool.Get()
}

func newRedisStorage(redisHost string, redisConnPoolSize int, redisPrefix string) *redisStorage {
	pool := newRedisConnectionPool(redisHost, redisConnPoolSize)
	return &redisStorage{pool, redisPrefix}
}

func (rs *redisStorage) Get(key string) (interface{}, error) {
	conn := rs.connection()
	defer conn.Close()
	data, err := redis.Bytes(conn.Do("GET", rs.prefix+key))
	if err == redis.ErrNil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return data, nil
}

func (rs *redisStorage) Set(key string, val interface{}) error {
	conn := rs.connection()
	defer conn.Close()
	var buffer = bytes.NewBuffer(nil)
	toStore := []byte{}
	if val != nil {
		enc := gob.NewEncoder(buffer)
		enc.Encode(val)
		toStore = buffer.Bytes()
	} else {
		toStore = nil
	}
	result, err := redis.String(conn.Do("SET", rs.prefix+key, toStore))
	if err != nil {
		return err
	}
	if result != "OK" {
		return errors.New("redis: SETEX call failed")
	}
	return nil
}

func (rs *redisStorage) Delete(key string) error {
	conn := rs.connection()
	defer conn.Close()
	result, err := redis.Int(conn.Do("DEL", rs.prefix+key))
	if err != nil {
		return err
	}
	if result != 1 {
		return errors.New("redis: DEL call failed")
	}
	return nil
}

func newRedisConnectionPool(host string, poolSize int) *redis.Pool {
	return redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", host)
		if err != nil {
			return nil, err
		}
		return c, err
	}, poolSize)
}
