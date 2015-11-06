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
	"reflect"
)

// LRCacheItem is returned from LRCache `Get` calls
type LRCacheItem struct {
	v         interface{}
	fromRedis bool
	err       error
}

func newLRCacheItem(v interface{}, err error, fromRedis bool) *LRCacheItem {
	return &LRCacheItem{
		v:         v,
		err:       err,
		fromRedis: fromRedis,
	}
}

// if there was an error fetching the value, this is where you'd find the description
func (l *LRCacheItem) Error() error {
	return l.err
}

// get raw value from cached item
func (l *LRCacheItem) Value() interface{} {
	return l.v
}

// get string value from cached item
func (l *LRCacheItem) StringVal() (string, error) {
	if l.Value() == nil {
		return "", nil
	}
	result := new(string)
	if l.IsGob() {
		buffer := l.Value().([]uint8)
		dec := gob.NewDecoder(bytes.NewBuffer(buffer))
		if err := dec.Decode(result); err != nil {
			return "", err
		} else {
			return *result, nil
		}
	} else {
		if reflect.TypeOf(l.Value()).String() == "string" {
			return l.Value().(string), nil
		} else {
			return "", errors.New("value is not a string")
		}
	}
}

// get int32 value from cached item
func (l *LRCacheItem) Int32Val() (int32, error) {
	if l.Value() == nil {
		return 0, nil
	}
	result := new(int32)
	if l.IsGob() {
		buffer := l.Value().([]uint8)
		dec := gob.NewDecoder(bytes.NewBuffer(buffer))
		if err := dec.Decode(result); err != nil {
			return 0, err
		} else {
			return *result, nil
		}
	} else {
		if reflect.TypeOf(l.Value()).String() == "int32" {
			return l.Value().(int32), nil
		} else {
			return 0, errors.New("value is not a int32")
		}
	}
}

// get int64 value from cached item
func (l *LRCacheItem) Int64Val() (int64, error) {
	if l.Value() == nil {
		return 0, nil
	}
	result := new(int64)
	if l.IsGob() {
		buffer := l.Value().([]uint8)
		dec := gob.NewDecoder(bytes.NewBuffer(buffer))
		if err := dec.Decode(result); err != nil {
			return 0, err
		} else {
			return *result, nil
		}
	} else {
		if reflect.TypeOf(l.Value()).String() == "int64" {
			return l.Value().(int64), nil
		} else {
			return 0, errors.New("value is not a int64")
		}
	}
}

// get int value from cached item
func (l *LRCacheItem) IntVal() (int, error) {
	if l.Value() == nil {
		return 0, nil
	}
	result := new(int)
	if l.IsGob() {
		buffer := l.Value().([]uint8)
		dec := gob.NewDecoder(bytes.NewBuffer(buffer))
		if err := dec.Decode(result); err != nil {
			return 0, err
		} else {
			return *result, nil
		}
	} else {
		if reflect.TypeOf(l.Value()).String() == "int" {
			return l.Value().(int), nil
		} else {
			return 0, errors.New("value is not a int")
		}
	}
}

// get float64 value from cached item
func (l *LRCacheItem) Float64Val() (float64, error) {
	if l.Value() == nil {
		return 0, nil
	}
	result := new(float64)
	if l.IsGob() {
		buffer := l.Value().([]uint8)
		dec := gob.NewDecoder(bytes.NewBuffer(buffer))
		if err := dec.Decode(result); err != nil {
			return 0, err
		} else {
			return *result, nil
		}
	} else {
		if reflect.TypeOf(l.Value()).String() == "float64" {
			return l.Value().(float64), nil
		} else {
			return 0, errors.New("value is not a float64")
		}
	}
}

func (l *LRCacheItem) FromRedis() bool {
	return l.fromRedis
}

// is the value a gob? true if the value was fetched from redis
func (l *LRCacheItem) IsGob() bool {
	rslt := false
	if l.v != nil && reflect.TypeOf(l.v).String() == "[]uint8" {
		rslt = true
	}
	return rslt
}
