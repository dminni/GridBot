package cache

import (
	"sync"
	"time"
)

type Item struct {
	Value      interface{}
	Expiration int64
}

type Cache struct {
	items sync.Map
}

func NewCache() *Cache {
	return &Cache{}
}

func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	var expiration int64
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}
	c.items.Store(key, Item{
		Value:      value,
		Expiration: expiration,
	})
}

func (c *Cache) Get(key string) (interface{}, bool) {
	item, ok := c.items.Load(key)
	if !ok {
		return nil, false
	}
	cached := item.(Item)
	if cached.Expiration > 0 && time.Now().UnixNano() > cached.Expiration {
		c.items.Delete(key)
		return nil, false
	}
	return cached.Value, true
}
