package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	Mu       sync.Mutex
	Entry    map[string]cacheEntry
	interval time.Duration
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(interval time.Duration) *Cache {
	c := &Cache {
		Mu:       sync.Mutex{},
		Entry:    make(map[string]cacheEntry),
		interval: interval,
	}
	go c.reapLoop()
	return c
}

// method that adds a new entry to the cache
func (c *Cache) Add(key string, val []byte) {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	c.Entry[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

// method that gets an entry from the cache
func (c *Cache) Get(key string) ([]byte, bool) {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	entry, ok := c.Entry[key]
	if !ok {
		return nil, false
	}
	return entry.val, true
}

// method that removes any entries that are older than the interval
func  (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)
	for range ticker.C {
		c.Mu.Lock()
		for key, entry := range c.Entry {
			if time.Since(entry.createdAt) > c.interval {
				delete(c.Entry, key)
			}
		}
		c.Mu.Unlock()
	}
}