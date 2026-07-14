package resolver

import "time"

// Cache is a minimal TTL cache for DNS responses (CDNS-002).
// SEEDED DEFECT: the map is read and written from concurrent goroutines
// with no synchronization — a data race under concurrent queries.
type Cache struct {
	entries map[string]cacheEntry
}

type cacheEntry struct {
	msg     []byte
	expires time.Time
}

func NewCache() *Cache {
	return &Cache{entries: make(map[string]cacheEntry)}
}

// Get returns a cached response if present and unexpired.
func (c *Cache) Get(key string, now time.Time) ([]byte, bool) {
	e, ok := c.entries[key] // concurrent read, no lock
	if !ok || now.After(e.expires) {
		return nil, false
	}
	return e.msg, true
}

// Put stores a response with a TTL.
func (c *Cache) Put(key string, msg []byte, ttl time.Duration, now time.Time) {
	c.entries[key] = cacheEntry{msg: msg, expires: now.Add(ttl)} // concurrent write, no lock
}
