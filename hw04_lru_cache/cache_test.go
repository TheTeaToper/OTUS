package hw04lrucache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})
}

func TestCacheCapacity(t *testing.T) {
	t.Run("Проверка переполнения кэша", func(t *testing.T) {
		cache := NewCache(4)
		cache.Set("1", "Меркурий")
		cache.Set("2", "Венера")
		cache.Set("3", "Земля")
		cache.Set("4", "Марс")
		cache.Set("5", "Уран")
		if _, mercuryFound := cache.Get("1"); mercuryFound {
			t.Error("Меркурий должен быть удален из кэша")
		}
	})
}

func TestCacheCleaning(t *testing.T) {
	t.Run("Проверка очистки кэша", func(t *testing.T) {
		cache := NewCache(3)
		cache.Set("One", 1)
		cache.Set("Two", 2)
		cache.Set("Three", 3)
		cache.Clear()
		_, firstFound := cache.Get("One")
		require.False(t, firstFound)
		_, secondFound := cache.Get("Two")
		require.False(t, secondFound)
		_, thirdFound := cache.Get("Three")
		require.False(t, thirdFound)
	})
}

func TestCacheMultithreading(t *testing.T) {
	t.Skip() // Remove me if task with asterisk completed.

	// c := NewCache(10)
	// wg := &sync.WaitGroup{}
	// wg.Add(2)

	// go func() {
	// 	defer wg.Done()
	// 	for i := 0; i < 1_000_000; i++ {
	// 		c.Set(Key(strconv.Itoa(i)), i)
	// 	}
	// }()

	// go func() {
	// 	defer wg.Done()
	// 	for i := 0; i < 1_000_000; i++ {
	// 		c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
	// 	}
	// }()

	// wg.Wait()
}
