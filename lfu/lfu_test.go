package lfu

import (
	"testing"

	"github.com/matryer/is"
)

func TestSet(t *testing.T) {
	is := is.New(t)

	cache := New(24, nil)
	cache.DelOldest()
	cache.Set("k1", 1)
	v := cache.Get("k1")
	is.Equal(v, 1)
	cache.Del("k1")
	is.Equal(0, cache.Len())
}

func TestOnEvict(t *testing.T) {
	is := is.New(t)

	keys := make([]string, 0, 8)
	onEvict := func(key string, value interface{}) {
		keys = append(keys, key)
	}
	cache := New(32, onEvict)

	cache.Set("k1", 1)
	cache.Set("k2", 2)
	cache.Set("k3", 3)
	cache.Set("k4", 4)

	expected := []string{"k1", "k3"}

	is.Equal(expected, keys)
	is.Equal(2, cache.Len())

}
