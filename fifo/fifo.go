package fifo

import (
	"container/list"

	"github.com/zhaoyi3264/cache"
)

type fifo struct {
	maxBytes int

	onEvict func(key string, value interface{})

	usedBytes int

	ll    *list.List
	cache map[string]*list.Element
}

type entry struct {
	key   string
	value interface{}
}

func (e *entry) Len() int {
	return cache.CalcLen(e.value)
}

func New(maxBytes int, onEvict func(key string, value interface{})) cache.Cache {
	return &fifo{
		maxBytes:  maxBytes,
		onEvict:   onEvict,
		usedBytes: 0,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
	}
}

func (f *fifo) Set(key string, value interface{}) {
	if e, ok := f.cache[key]; ok {
		f.ll.MoveToBack(e)
		en := e.Value.(*entry)
		f.usedBytes = f.usedBytes - cache.CalcLen(en.value) + cache.CalcLen(value)
		en.value = value
	} else {
		en := &entry{key, value}
		e := f.ll.PushBack(en)
		f.cache[key] = e

		f.usedBytes += en.Len()
		if f.maxBytes > 0 && f.usedBytes > f.maxBytes {
			f.DelOldest()
		}
	}
}

func (f *fifo) Get(key string) interface{} {
	if e, ok := f.cache[key]; ok {
		return e.Value.(*entry).value
	}
	return nil
}

func (f *fifo) Del(key string) {
	if e, ok := f.cache[key]; ok {
		f.removeElement(e)
	}
}

func (f *fifo) DelOldest() {
	f.removeElement(f.ll.Front())
}

func (f *fifo) removeElement(e *list.Element) {
	if e == nil {
		return
	}
	f.ll.Remove(e)
	en := e.Value.(*entry)
	f.usedBytes -= en.Len()
	delete(f.cache, en.key)
	if f.onEvict != nil {
		f.onEvict(en.key, en.value)
	}
}

func (f *fifo) Len() int {
	return f.ll.Len()
}
