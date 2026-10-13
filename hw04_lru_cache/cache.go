package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type cacheItem struct {
	key   Key
	value interface{}
}

type lruCache struct {
	Cache // Remove me after realization.

	capacity int
	queue    List
	items    map[Key]*ListItem
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (cache *lruCache) Set(key Key, value interface{}) bool {
	if existingItem, ok := cache.items[key]; ok {
		existingItem.Value = cacheItem{key, value}
		cache.queue.MoveToFront(existingItem)
		return true
	}

	newItem := cache.queue.PushFront(cacheItem{key, value})
	cache.items[key] = newItem
	if cache.queue.Len() > cache.capacity {
		lastItem := cache.queue.Back()
		delete(cache.items, lastItem.Value.(cacheItem).key)
		cache.queue.Remove(lastItem)
	}
	return false
}

func (cache *lruCache) Get(key Key) (interface{}, bool) {
	if existingItem, ok := cache.items[key]; ok {
		cache.queue.MoveToFront(existingItem)
		return existingItem.Value.(cacheItem).value, true
	}
	return nil, false
}

func (cache *lruCache) Clear() {
	cache.queue = NewList()
	cache.items = make(map[Key]*ListItem, cache.capacity)
}
