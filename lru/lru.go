package lru

import "container/list"

type Cache struct {
	maxBytes int64
	nowBytes int64

	ll    *list.List
	cache map[string]*list.Element

	OnEvicted func(key string, value Value)
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     map[string]*list.Element{},
		OnEvicted: onEvicted,
	}
}

// k v
type entry struct {
	key   string
	value Value
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

func (c *Cache) Get(key string) (Value, bool) {
	if ele, ok := c.cache[key]; ok {
		kv := ele.Value.(*entry)
		c.ll.MoveToFront(ele)
		return kv.value, true
	}
	return nil, false
}

func (c *Cache) RemoveOldest() {
	tail := c.ll.Back()
	c.ll.Remove(tail)
	kv := tail.Value.(*entry)
	delete(c.cache, kv.key)
	c.nowBytes -= int64(len(kv.key) + kv.value.Len())
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

func (c *Cache) Add(key string, value Value) {
	// 已存在, 移到队首， 更新值
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nowBytes += int64(kv.value.Len() - value.Len())
		kv.value = value
		return
	} else {
		element := c.ll.PushFront(&entry{
			key:   key,
			value: value,
		})
		c.cache[key] = element
		c.nowBytes += int64(len(key) + value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nowBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
