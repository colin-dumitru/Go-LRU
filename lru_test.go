package lru

import (
	"testing"
)

func TestEmptyCache(t *testing.T) {
	cache, _ := New(128, nil)

	if cache.EmptySpace() != 128 {
		t.Fail()
	}
}

func TestBasicTest_NoProducer(t *testing.T) {
	cache, _ := New(128, nil)

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
	})

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
	})

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
	})

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
	cache, _ := New(128, nil)

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
	cache, _ := New(128, nil)

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
