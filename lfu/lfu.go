package lfu

import (
	"container/heap"

	"github.com/zhaoyi3264/cache"
)

type lfu struct {
	maxBytes  int
	onEvict   func(key string, value interface{})
	usedBytes int
	queue     *queue
	cache     map[string]*entry
}

type entry struct {
	key    string
	value  interface{}
	weight int
	index  int
}

func (e *entry) Len() int {
	return cache.CalcLen(e.value) + 4 + 4
}

type queue []*entry

func (q queue) Len() int {
	return len(q)
}

func (q queue) Less(i, j int) bool {
	return q[i].weight < q[j].weight
}

func (q queue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

func (q *queue) Push(x interface{}) {
	n := len(*q)
	en := x.(*entry)
	en.index = n
	*q = append(*q, en)
}

func (q *queue) Pop() interface{} {
	old := *q
	n := len(old)
	en := old[n-1]
	old[n-1] = nil
	en.index = -1
	*q = old[0 : n-1]
	return en
}

func New(maxBytes int, onEvict func(key string, value interface{})) cache.Cache {
	q := make(queue, 0, 1024)
	return &lfu{
		maxBytes: maxBytes,
		onEvict:  onEvict,
		queue:    &q,
		cache:    make(map[string]*entry),
	}
}

func (l *lfu) Set(key string, value interface{}) {
	if e, ok := l.cache[key]; ok {
		l.usedBytes = l.usedBytes - cache.CalcLen(e.value) + cache.CalcLen(value)
		l.queue.update(e, value, e.weight+1)
	} else {
		en := &entry{key: key, value: value}
		heap.Push(l.queue, en)
		l.cache[key] = en

		l.usedBytes += en.Len()
		if l.maxBytes > 0 && l.usedBytes > l.maxBytes {
			l.removeElement(heap.Pop(l.queue))
		}
	}
}

func (q *queue) update(en *entry, value interface{}, weight int) {
	en.value = value
	en.weight = weight
	heap.Fix(q, en.index)
}

func (l *lfu) Get(key string) interface{} {
	if e, ok := l.cache[key]; ok {
		l.queue.update(e, e.value, e.weight+1)
		return e.value
	}
	return nil
}

func (l *lfu) Del(key string) {
	if e, ok := l.cache[key]; ok {
		heap.Remove(l.queue, e.index)
		l.removeElement(e)
	}
}

func (l *lfu) DelOldest() {
	if l.queue.Len() == 0 {
		return
	}
	l.removeElement(heap.Pop(l.queue))
}

func (l *lfu) removeElement(x interface{}) {
	if x == nil {
		return
	}
	en := x.(*entry)

	delete(l.cache, en.key)

	l.usedBytes -= en.Len()
	if l.onEvict != nil {
		l.onEvict(en.key, en.value)
	}
}

func (l *lfu) Len() int {
	return l.queue.Len()
}
