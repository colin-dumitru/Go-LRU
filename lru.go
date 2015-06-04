/**
 * Package for managing LRU caching for items of arbitrary sizes. Most
 * existing LRU cache implementation have a maximum item size, while this specific
 * version defines a maximum cache size with items of arbitrary dimensions. Usefull
 * for storing blob data (files, text etc.) with a maximum memory usage.
 */

package lru

import (
	"container/list"
	"errors"
	"sync"
)

/**
 * LRU cache used for storing items of arbitrary size.
 */
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
	producer func(key interface{}) *LRUItem

	// Mutex used to synchronise cache operations
	rwlock sync.RWMutex
}

/**
 * Single cache item
 */
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

/**
 * Creates a new LRW cache
 * @param {maxSize} - the maximum size of this cache
 * @param {producer} - a callback function which gets invoked on cache misses. The
 * 	result of this callback is stored in the cache for the key which was missing.
 * 	Can be null, in which case all cache misses will return null
 */
func New(maxSize int, producer func(key interface{}) *LRUItem) (*LRUCache, error) {
	if maxSize <= 0 {
		return nil, errors.New("Cache size must be greater than 0")
	}

	cache := &LRUCache{
		maxSize:     maxSize,
		currentSize: 0,

		orderedItems: list.New(),
		items:        make(map[interface{}]*LRUItem),

		producer: producer,
	}
	return cache, nil
}

/**
 * Retrieves an element from the cache. In case of cache miss and a producer is defined, then
 * the producer is invoked and the result is stored in the cache. If the producer is null, no
 * result is returned (nil return value)
 *
 * @param key - the key to search in the cache
 * @return - if found, returns the corresponding cached element. If not found, the result
 * 	can either be null if the producer was not given, or the retult of the producer
 */
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
}

func (cache *LRUCache) Put(key interface{}, element *LRUItem) error {
	cache.rwlock.Lock()
	defer cache.rwlock.Unlock()

	if _, ok := cache.items[key]; ok {
		return errors.New("Key already exists")
	}
	cache.putItem(&key, element)
	return nil
}

func (cache *LRUCache) EmptySpace() int {
	cache.rwlock.RLock()
	defer cache.rwlock.RUnlock()

	return cache.maxSize - cache.currentSize
}

func (cache *LRUCache) Has(key interface{}) bool {
	cache.rwlock.RLock()
	defer cache.rwlock.RUnlock()

	_, ok := cache.items[key]

	return ok
}
