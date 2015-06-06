package lru

import (
	"testing"
)

func TestEmptyCache(t *testing.T) {
	cache, _ := New(128, nil, nil)

	if cache.EmptySpace() != 128 {
		t.Fail()
	}
}

func TestBasicTest_NoProducer(t *testing.T) {
	cache, _ := New(128, nil, nil)

	elem := cache.Get("test")

	if elem != nil {
		t.Fail()
	}
}

func TestBasicTest_Producer(t *testing.T) {
	cache, _ := New(128, func(key interface{}) *LRUItem {
		if key == "test" {
			return &LRUItem{
				Value: "test-value",
				Size:  100,
			}
		}
		return nil
	}, nil)

	elem := cache.Get("test")

	if elem.Value != "test-value" {
		t.Fail()
	}

	elem2 := cache.Get("test")

	if elem2.Value != "test-value" {
		t.Fail()
	}
	if cache.EmptySpace() != 28 {
		t.Fail()
	}
}

func TestBasicTest_Eviction(t *testing.T) {
	cache, _ := New(128, func(key interface{}) *LRUItem {
		return &LRUItem{
			Value: key,
			Size:  50,
		}
	}, nil)

	cache.Get("test1")
	cache.Get("test2")
	cache.Get("test3")

	if cache.Has("test1") {
		t.Fail()
	}
	if !cache.Has("test2") {
		t.Fail()
	}
	if !cache.Has("test3") {
		t.Fail()
	}
}

func TestBasicTest_Eviction2(t *testing.T) {
	cache, _ := New(128, func(key interface{}) *LRUItem {
		return &LRUItem{
			Value: key,
			Size:  50,
		}
	}, nil)

	cache.Get("test1")
	cache.Get("test2")
	cache.Get("test1")
	cache.Get("test3")

	if !cache.Has("test1") {
		t.Fail()
	}
	if cache.Has("test2") {
		t.Fail()
	}
	if !cache.Has("test3") {
		t.Fail()
	}
}

func TestBasicTest_PutAndEviction(t *testing.T) {
	cache, _ := New(128, nil, nil)

	if cache.Get("test1") != nil {
		t.Fail()
	}

	cache.Put("test1", &LRUItem{
		Value: "test1",
		Size:  50,
	})
	cache.Put("test2", &LRUItem{
		Value: "test2",
		Size:  50,
	})
	cache.Get("test1")
	cache.Put("test3", &LRUItem{
		Value: "test3",
		Size:  50,
	})
	cache.Get("test1")
	cache.Put("test4", &LRUItem{
		Value: "test4",
		Size:  50,
	})

	if !cache.Has("test1") {
		t.Fail()
	}
	if cache.Has("test2") {
		t.Fail()
	}
	if cache.Has("test3") {
		t.Fail()
	}
	if !cache.Has("test4") {
		t.Fail()
	}
}

func TestBasicTest_Has(t *testing.T) {
	cache, _ := New(128, nil, nil)

	if cache.Has("test1") {
		t.Fail()
	}

	cache.Put("test1", &LRUItem{
		Value: "test1",
		Size:  50,
	})

	if !cache.Has("test1") {
		t.Fail()
	}
}

func TestOnEvictHandler(t *testing.T) {
	itemsEvicted := ""

	cache, _ := New(128, nil, func(key interface{}, item *LRUItem) {
		itemsEvicted += key.(string)
	})

	cache.Put("test1", &LRUItem{
		Value: "test1",
		Size:  50,
	})
	cache.Put("test2", &LRUItem{
		Value: "test2",
		Size:  50,
	})

	if itemsEvicted != "" {
		t.Fail()
	}
	cache.Put("test3", &LRUItem{
		Value: "test3",
		Size:  100,
	})
	if itemsEvicted != "test1test2" {
		t.Fail()
	}
}

func TestMakeRoom(t *testing.T) {
	itemsEvicted := ""

	cache, _ := New(160, nil, func(key interface{}, item *LRUItem) {
		itemsEvicted += key.(string)
	})

	cache.Put("test1", &LRUItem{
		Value: "test1",
		Size:  50,
	})
	cache.Put("test2", &LRUItem{
		Value: "test2",
		Size:  50,
	})

	if itemsEvicted != "" {
		t.Fail()
	}

	cache.MakeRoom(100)

	if itemsEvicted != "test1" {
		t.Fail()
	}
	if cache.EmptySpace() != 110 {
		t.Fail()
	}

	cache.Put("test3", &LRUItem{
		Value: "test3",
		Size:  100,
	})
	if itemsEvicted != "test1" {
		t.Fail()
	}
}

func TestPut_ItemExists(t *testing.T) {
	cache, _ := New(100, nil, nil)

	cache.Put("test1", &LRUItem{
		Value: "test1",
		Size:  50,
	})
	error := cache.Put("test1", &LRUItem{
		Value: "test2",
		Size:  50,
	})

	if error == nil {
		t.Fail()
	}

	item := cache.Get("test1")

	if item.Value != "test1" {
		t.Fail()
	}
}

func TestReplace_ItemExists(t *testing.T) {
	cache, _ := New(100, nil, nil)

	cache.Put("test1", &LRUItem{
		Value: "test1",
		Size:  50,
	})
	cache.Replace("test1", &LRUItem{
		Value: "test2",
		Size:  50,
	})

	item := cache.Get("test1")

	if item.Value != "test2" {
		t.Fail()
	}
}

func TestReplace_ItemNotExists(t *testing.T) {
	cache, _ := New(100, nil, nil)

	cache.Replace("test1", &LRUItem{
		Value: "test1",
		Size:  50,
	})
	cache.Replace("test2", &LRUItem{
		Value: "test2",
		Size:  50,
	})

	item := cache.Get("test2")

	if item.Value != "test2" {
		t.Fail()
	}
}

func TestEvict_ItemExists(t *testing.T) {
	cache, _ := New(100, nil, nil)

	cache.Replace("test1", &LRUItem{
		Value: "test1",
		Size:  50,
	})
	evicted := cache.Evict("test1")

	if evicted == nil {
		t.Fail()
	}

	item := cache.Get("test1")

	if item != nil {
		t.Fail()
	}
}

func TestEvict_ItemNotExists(t *testing.T) {
	cache, _ := New(100, nil, nil)

	evicted := cache.Evict("test1")

	if evicted != nil {
		t.Fail()
	}
}

func TestEvictIf(t *testing.T) {
	cache, _ := New(100, nil, nil)

	cache.Put("test1", &LRUItem{Value: "test", Size: 10})
	cache.Put("test2", &LRUItem{Value: "test", Size: 10})
	cache.Put("test3", &LRUItem{Value: "test", Size: 10})

	cache.EvictIf(func(key interface{}) bool {
		return key == "test2"
	})

	if !cache.Has("test1") {
		t.Fail()
	}
	if cache.Has("test2") {
		t.Fail()
	}
	if !cache.Has("test3") {
		t.Fail()
	}

}

func TestEvictIf_AllElements(t *testing.T) {
	cache, _ := New(100, nil, nil)

	cache.Put("test1", &LRUItem{Value: "test", Size: 10})
	cache.Put("test2", &LRUItem{Value: "test", Size: 10})
	cache.Put("test3", &LRUItem{Value: "test", Size: 10})

	cache.EvictIf(func(key interface{}) bool {
		return true
	})

	if cache.Has("test1") {
		t.Fail()
	}
	if cache.Has("test2") {
		t.Fail()
	}
	if cache.Has("test3") {
		t.Fail()
	}
}

func TestEvictIf_NoElements(t *testing.T) {
	cache, _ := New(100, nil, nil)

	cache.Put("test1", &LRUItem{Value: "test", Size: 10})
	cache.Put("test2", &LRUItem{Value: "test", Size: 10})
	cache.Put("test3", &LRUItem{Value: "test", Size: 10})

	cache.EvictIf(func(key interface{}) bool {
		return false
	})

	if !cache.Has("test1") {
		t.Fail()
	}
	if !cache.Has("test2") {
		t.Fail()
	}
	if !cache.Has("test3") {
		t.Fail()
	}
}
