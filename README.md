# LRCache [![Circle CI](https://circleci.com/gh/streamrail/lrcache.svg?style=svg)](https://circleci.com/gh/streamrail/lrcache) [![GoDoc](https://godoc.org/github.com/streamrail/lrcache?status.svg)](https://godoc.org/github.com/streamrail/lrcache)

## Summary

`LRCache` aims to provide a simple yet fast and scalable two-layer caching solution. It exposes only simple `Set` and `Get` functions that cache items in-memory in an `LRU-cache` implementation (using [golang-lru](https://github.com/hashicorp/golang-lru)). Hot items are fetched from the in-memory `LRU-cache`, while less-hot items that are evicted from `LRU-cache` are fetched from `Redis` (using [Redigo](https://github.com/garyburd/redigo)). This offers speed on one hand (hot items are grabbed from the in-memory `LRU-cache`) and scalability on the other hand (large data sets may be cached using a `Redis` server or cluster).

## Usage

Example (see [example/simple/main.go](https://github.com/streamrail/lrcache/blob/master/example/simple/main.go) for full runnable example): Set 10 elements in the cache. Since the `LRU-cache` size is set to 5 with a flag, we expect 5 elements in the `LRU-cache`, 5 evictions from `LRU-cache`, and 5 elements in `Redis`. 

```go
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

---- output ---- 
setting key 0 with value 0. evicted stuff from cache? false
setting key 1 with value 1. evicted stuff from cache? false
setting key 2 with value 2. evicted stuff from cache? false
setting key 3 with value 3. evicted stuff from cache? false
setting key 4 with value 4. evicted stuff from cache? false
setting key 5 with value 5. evicted stuff from cache? true
setting key 6 with value 6. evicted stuff from cache? true
setting key 7 with value 7. evicted stuff from cache? true
setting key 8 with value 8. evicted stuff from cache? true
setting key 9 with value 9. evicted stuff from cache? true
value came from redis: 0
value came from redis: 1
value came from redis: 2
value came from redis: 3
value came from redis: 4
value came from LRU: 5
value came from LRU: 6
value came from LRU: 7
value came from LRU: 8
value came from LRU: 9
```

## Using the cache with custom types (`interface` vs `gob`)

If you you know what type you have set in the cache, use the appropriate value getter (e.g. `Int32Val()`, `StringVal()`, etc.). Otherwise or if you have set your own custom type in the cache, you can get the raw value and parse it yourself. But, in this case, you would have to use the `FromRedis()` function (or the `IsGob()` function). This is because, before they could be stored on `Redis`, objects have to be  serialized to `gobs`, and when calling the `Get(key)` function, you do not know if the value was returned from the `LRU-cache` (where it is stored as an `interface` in memory) or from `Redis` (where it is stored as a gob) unless you call `FromRedis()`. Also, `LRCache` cannot perform the deserialization from `gob` to the custom `type` for you, since the `gob decoder` cannot decode a `gob` value to an `interface`. 

Example: extract custom `type` from cache (see [example/custom_type/main.go](https://github.com/streamrail/lrcache/blob/master/example/custom_type/main.go) for full runnable example):

```go

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
	if !res.FromRedis() { // value from `LRU-cache`
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

---- output ---- 
from Redis cache:
&{Bob 21} <nil>
from LRU cache:
&{Alice 19} <nil>

``` 

## Pitfall when using `Set` with `LRCache`

When working with a "pure" `LRU-cache`, once an item gets evicted, you could not find that item in the cache anymore. This has thd downside of a cache miss, but it also has the upside of forcing you to re-insert the item into the `LRU-cache` in case you want it there. Re-insertions are common with `LRU-cache`. This constant re-insertion to the `LRU-cache` means that if an items was evicted in the past, yet suddenly it becomes a hot item that's been re-inserted and is now requested a lot, it will be returned from the in-memory `LRU-cache`. 

In contrast, on `LRCache`, when you set an item with `Set`, there is more seldom a cache-miss because of an LRU eviction. If the `LRU-cache` evicted a cold item, that item may still be stored on the `Redis` cache and the `Get` method may be able to resolve it. This has the upside of good performance for all cache `Get` requests (there is never cache-miss as in the pure `LRU-cache` example above), however - the price we pay for this is that the in-memory `LRU-cache` component becomes less adaptive. 

If an item that was evicted from `LRU-cache` is still returned from `Redis`, you won't need to re-insert the item to the in-memory `LRU-cache` (you can do it if you like, but `LRCache` won't do it for you). This means that an item that turned cold and got evicted from the `LRU-cache` may sometime in the future become hot again, yet it will still be returned from `Redis` (unless you explicitly call the `Set` method for this item) and not in-memory `LRU-cache` - until the item gets evicted from `Redis` as well. 

## `Redis` config

For the reason depicted in the pitfall section, it is advised to use `LRCache` with `Redis` configured as an `LRU-cache` itself. Read more about [using Redis as an `LRU-Cache`](http://redis.io/topics/lru-cache) on the official `Redis` website.

This configuration works to guarantee that a cache-miss will be generated after an item is truly cold (i.e. not requested from neither the in-memory `LRU-cache` component nor the `Redis` component of `LRCache`).


## License
MIT (see [LICENSE](https://github.com/streamrail/lrcache/blob/master/LICENSE) file)
