package lru


import "container/list"

// 缓存结构体
type Cache struct {
	maxBytes int64 //最大容量
	nbytes   int64 //当前使用容量
	ll       *list.List
	cache    map[string]*list.Element
	OnEvicted func(key string, value Value) //回调函数
}
//节点
type entry struct {
	key   string
	value Value
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

// New is the Constructor of Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Add adds a value to the cache.
func (c *Cache) Add(key string, value Value) {
	/*
	首先查找这个缓存是否存在：
	1.存在：更新缓存，并将缓存移到队尾
	2.不存在：将数据缓存到队尾
	 */
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	//检测缓存是否达到最大容量，如果达到将删除队首
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Get look ups a key's value
func (c *Cache) Get(key string) (value Value, ok bool) {
	/*
	缓存中查找数据：
	1。存在：将数据移到队尾，返回
	2。不存在：返回空数据
	 */
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest removes the oldest item
func (c *Cache) RemoveOldest() {
	/*
	删除元素流程：
	1。取出队首
	2。将队首断言成entry
	3。删除队首
	4。删除map中缓存
	5。更新当前使用内存
	6。运行回调函数
	 */
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Len the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}