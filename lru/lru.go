package lru

import (
	"container/list"

	"github.com/zhaoyi3264/cache"
)

type lru struct {
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
	return &lru{
		maxBytes:  maxBytes,
		onEvict:   onEvict,
		usedBytes: 0,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
	}
}

func (l *lru) Set(key string, value interface{}) {
	if e, ok := l.cache[key]; ok {
		l.ll.MoveToBack(e)
		en := e.Value.(*entry)
		l.usedBytes = l.usedBytes - cache.CalcLen(en.value) + cache.CalcLen(value)
		en.value = value
	} else {
		en := &entry{key, value}
		e := l.ll.PushBack(en)
		l.cache[key] = e

		l.usedBytes += en.Len()
		if l.maxBytes > 0 && l.usedBytes > l.maxBytes {
			l.DelOldest()
		}
	}
}

func (l *lru) Get(key string) interface{} {
	if e, ok := l.cache[key]; ok {
		l.ll.MoveToBack(e)
		return e.Value.(*entry).value
	}
	return nil
}

func (l *lru) Del(key string) {
	if e, ok := l.cache[key]; ok {
		l.removeElement(e)
	}
}

func (l *lru) DelOldest() {
	l.removeElement(l.ll.Front())
}

func (l *lru) removeElement(e *list.Element) {
	if e == nil {
		return
	}
	l.ll.Remove(e)
	en := e.Value.(*entry)
	l.usedBytes -= en.Len()
	delete(l.cache, en.key)
	if l.onEvict != nil {
		l.onEvict(en.key, en.value)
	}
}

func (l *lru) Len() int {
	return l.ll.Len()
}
