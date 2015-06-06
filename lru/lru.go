// Package lru contains an implementation for LRU caching with items of arbitrary sizes. Most
// existing LRU cache implementation have a maximum item size, while this specific
// version defines a maximum cache size with items of arbitrary dimensions. Usefull
// for storing blob data (files, text etc.) with a maximum memory usage.
package lru

import (
	"container/list"
	"errors"
	"sync"
)

// LRU cache used for storing items of arbitrary size.
type LRUCache struct {
	// Maximum size the cache can hold
	maxSize int
	// Current occupied size
	currentSize int

	// Cached items sorted by the last time they were read. The last item(s)
	// in this queue will be evicted if no more space is left
	orderedItems *list.List
	// Map of item key to value
	items map[interface{}]*LRUItem

	// In case of cache-mises, this producer will create the item to be stored in the cache
	producer func(interface{}) *LRUItem
	// Callback when an item is evicted from the cache
	onevict func(interface{}, *LRUItem)

	// Mutex used to synchronise cache operations
	rwlock sync.RWMutex
}

// Single cache item
type LRUItem struct {
	// Actual value stored for this specific item
	Value interface{}
	// The size for this item
	Size int

	// The list element stored in the "orderedList" list for this item. Used for faster
	// cache manipulation
	elementListItem *list.Element
	// The key for this cache, as it's stored within the "items" map
	elementKey *interface{}
}

// New Creates a new LRW cache
// @param {maxSize} - the maximum size of this cache
// @param {producer} - optional callback function which gets invoked on cache misses. The
// 	result of this callback is stored in the cache for the key which was missing.
// 	Can be null, in which case all cache misses will return null
// @param {onevict} - optional function which will get called when an item is evicted from the cache
func New(maxSize int, producer func(key interface{}) *LRUItem, onevict func(interface{}, *LRUItem)) (*LRUCache, error) {
	if maxSize <= 0 {
		return nil, errors.New("Cache size must be greater than 0")
	}

	cache := &LRUCache{
		maxSize:     maxSize,
		currentSize: 0,

		orderedItems: list.New(),
		items:        make(map[interface{}]*LRUItem),

		producer: producer,
		onevict:  onevict,
	}
	return cache, nil
}

// Get retrieves an element from the cache. In case of cache miss and a producer is defined, then
// the producer is invoked and the result is stored in the cache. If the producer is null, no
// result is returned (nil return value)
//
// @param key - the key to search in the cache
// @return - if found, returns the corresponding cached element. If not found, the result
// 	can either be null if the producer was not given, or the retult of the producer
func (cache *LRUCache) Get(key interface{}) *LRUItem {
	cache.rwlock.Lock()
	defer cache.rwlock.Unlock()

	if element, ok := cache.items[key]; ok {
		cache.orderedItems.MoveToFront(element.elementListItem)
		return element
	} else {
		return cache.produceItem(key)
	}
}

func (cache *LRUCache) produceItem(key interface{}) *LRUItem {
	if cache.producer == nil {
		return nil
	}

	element := cache.producer(key)

	if element == nil {
		return nil
	}

	cache.putItem(&key, element)
	return element
}

func (cache *LRUCache) putItem(key *interface{}, element *LRUItem) {
	cache.currentSize += element.Size

	listElement := cache.orderedItems.PushFront(element)
	cache.items[*key] = element

	element.elementListItem = listElement
	element.elementKey = key

	cache.evictAsNeeded()
}

func (cache *LRUCache) evictAsNeeded() {
	for cache.currentSize > cache.maxSize && cache.orderedItems.Len() > 0 {
		back := cache.orderedItems.Back()
		cache.evictElement(back.Value.(*LRUItem))
	}
}

func (cache *LRUCache) evictElement(element *LRUItem) {
	cache.currentSize -= element.Size

	delete(cache.items, *element.elementKey)
	cache.orderedItems.Remove(element.elementListItem)

	if cache.onevict != nil {
		cache.onevict(*element.elementKey, element)
	}
}

// MakeRoom evicts elements from the cache until the specified empty space is made. Ff the
// cache already has enough empty space, then no elements are evicted.
// @param  {int} size - how much empty space should be ensured
func (cache *LRUCache) MakeRoom(size int) {
	cache.rwlock.Lock()
	defer cache.rwlock.Unlock()

	cache.maxSize -= size
	cache.evictAsNeeded()
	cache.maxSize += size
}

// Put adds a new element to the cache. If an item with the same key already exists,
// then the operation fails and an error is returned.
// @param {interface{}} key - the key for the new cached item
// @param {*LRUItem} element - the element to be inserted
func (cache *LRUCache) Put(key interface{}, element *LRUItem) error {
	cache.rwlock.Lock()
	defer cache.rwlock.Unlock()

	if _, ok := cache.items[key]; ok {
		return errors.New("Key already exists")
	}
	cache.putItem(&key, element)
	return nil
}

// Replace adds a new element to the cache. If an item with the same key already exists,
// then it is evicted and replaced with the given element.
// @param {interface{}} key - the key for the new cached item
// @param {*LRUItem} element - the element to be inserted
func (cache *LRUCache) Replace(key interface{}, element *LRUItem) error {
	cache.rwlock.Lock()
	defer cache.rwlock.Unlock()

	if item, ok := cache.items[key]; ok {
		cache.evictElement(item)
	}
	cache.putItem(&key, element)
	return nil
}

// Evict removes an element from the cache.
// @param {interface{}} key - the key of the element to be removed
// @return {*LRUItem} - if an element with the specified key is found and removed, then the return
// value is the deleted element. Otherwise, nil is returned.
func (cache *LRUCache) Evict(key interface{}) *LRUItem {
	cache.rwlock.Lock()
	defer cache.rwlock.Unlock()

	if item, ok := cache.items[key]; ok {
		cache.evictElement(item)
		return item
	}
	return nil
}

// Evict removes an element from the cache.
// @param {interface{}} key - the key of the element to be removed
// @return {*LRUItem} - if an element with the specified key is found and removed, then the return
// value is the deleted element. Otherwise, nil is returned.
func (cache *LRUCache) EvictIf(predicate func(interface{}) bool) *LRUItem {
	cache.rwlock.Lock()
	defer cache.rwlock.Unlock()

	for key, value := range cache.items {
		if predicate(key) {
			cache.evictElement(value)
		}
	}
	return nil
}

// EmptySpace returns the remaining empty space within the cache.
func (cache *LRUCache) EmptySpace() int {
	cache.rwlock.RLock()
	defer cache.rwlock.RUnlock()

	return cache.maxSize - cache.currentSize
}

// MaxSize returns the maximum size this cache can hold.
func (cache *LRUCache) MaxSize() int {
	cache.rwlock.RLock()
	defer cache.rwlock.RUnlock()

	return cache.maxSize
}

// Has checks if an element with the specified key exists within the cache
// @param {interface{}} key - the key of the element
// @return {bool} - if an item with the specified key exists
func (cache *LRUCache) Has(key interface{}) bool {
	cache.rwlock.RLock()
	defer cache.rwlock.RUnlock()

	_, ok := cache.items[key]

	return ok
}
