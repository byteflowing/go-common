package cache

import (
	"errors"

	"github.com/coocood/freecache"
)

var (
	ErrNotFound = errors.New("not found")
)

type Cache struct {
	cli *freecache.Cache
}

type Opts struct {
	Size int // 缓存容量 单位：bytes
}

func New(opts *Opts) *Cache {
	return &Cache{
		cli: freecache.NewCache(opts.Size),
	}
}

// Set sets a key, value and expiration for a cache entry and stores it in the cache.
// If the key is larger than 65535 or value is larger than 1/ 1024 of the cache size,
// the entry will not be written to the cache.
// expireSeconds <= 0 means no expire, but it can be evicted when cache is full
func (c *Cache) Set(key string, value []byte, expireSeconds int) (err error) {
	return c.cli.Set([]byte(key), value, expireSeconds)
}

// Get 获取key对应的值
// 如果没有找到返回ErrNotFound
func (c *Cache) Get(key string) (value []byte, err error) {
	value, err = c.cli.Get([]byte(key))
	if err != nil {
		if errors.Is(err, freecache.ErrNotFound) {
			err = ErrNotFound
		}
	}
	return
}

// Delete 删除key指定的值
func (c *Cache) Delete(key string) {
	c.cli.Del([]byte(key))
}
