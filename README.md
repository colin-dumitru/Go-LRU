# lru
--
    import "."

Package lru contains an implementation for LRU caching with items of arbitrary
sizes. Most existing LRU cache implementation have a maximum item size, while
this specific version defines a maximum cache size with items of arbitrary
dimensions. Usefull for storing blob data (files, text etc.) with a maximum
memory usage.

## Usage

#### type LRUCache

```go
type LRUCache struct {
}
```

LRU cache used for storing items of arbitrary size.

#### func  New

```go
func New(maxSize int, producer func(key interface{}) *LRUItem, onevict func(interface{}, *LRUItem)) (*LRUCache, error)
```
New Creates a new LRW cache @param {maxSize} - the maximum size of this cache
@param {producer} - optional callback function which gets invoked on cache
misses. The

    result of this callback is stored in the cache for the key which was missing.
    Can be null, in which case all cache misses will return null

@param {onevict} - optional function which will get called when an item is
evicted from the cache

#### func (*LRUCache) EmptySpace

```go
func (cache *LRUCache) EmptySpace() int
```
EmptySpace returns the remaining empty space within the cache.

#### func (*LRUCache) Evict

```go
func (cache *LRUCache) Evict(key interface{}) *LRUItem
```
Evict removes an element from the cache. @param {interface{}} key - the key of
the element to be removed @return {*LRUItem} - if an element with the specified
key is found and removed, then the return value is the deleted element.
Otherwise, nil is returned.

#### func (*LRUCache) Get

```go
func (cache *LRUCache) Get(key interface{}) *LRUItem
```
Get retrieves an element from the cache. In case of cache miss and a producer is
defined, then the producer is invoked and the result is stored in the cache. If
the producer is null, no result is returned (nil return value)

@param key - the key to search in the cache @return - if found, returns the
corresponding cached element. If not found, the result

    can either be null if the producer was not given, or the retult of the producer

#### func (*LRUCache) Has

```go
func (cache *LRUCache) Has(key interface{}) bool
```
Has checks if an element with the specified key exists within the cache @param
{interface{}} key - the key of the element @return {bool} - if an item with the
specified key exists

#### func (*LRUCache) MakeRoom

```go
func (cache *LRUCache) MakeRoom(size int)
```
MakeRoom evicts elements from the cache until the specified empty space is made.
Ff the cache already has enough empty space, then no elements are evicted.
@param {int} size - how much empty space should be ensured

#### func (*LRUCache) MaxSize

```go
func (cache *LRUCache) MaxSize() int
```
MaxSize returns the maximum size this cache can hold.

#### func (*LRUCache) Put

```go
func (cache *LRUCache) Put(key interface{}, element *LRUItem) error
```
Put adds a new element to the cache. If an item with the same key already
exists, then the operation fails and an error is returned. @param {interface{}}
key - the key for the new cached item @param {*LRUItem} element - the element to
be inserted

#### func (*LRUCache) Replace

```go
func (cache *LRUCache) Replace(key interface{}, element *LRUItem) error
```
Replace adds a new element to the cache. If an item with the same key already
exists, then it is evicted and replaced with the given element. @param
{interface{}} key - the key for the new cached item @param {*LRUItem} element -
the element to be inserted

#### type LRUItem

```go
type LRUItem struct {
	// Actual value stored for this specific item
	Value interface{}
	// The size for this item
	Size int
}
```

Single cache item
