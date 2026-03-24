package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

var ctx = context.Background()

func TestLocalCache_SetAndGet(t *testing.T) {
	c, _ := NewLocalCache(10)
	defer c.Close()

	c.Set(ctx, "key", []byte("value"), time.Minute)

	val, err := c.Get(ctx, "key")
	if err != nil {
		t.Fatalf("esperaba valor, got error: %v", err)
	}
	if string(val) != "value" {
		t.Fatalf("esperaba 'value', got '%s'", val)
	}
}

func TestLocalCache_Expiration(t *testing.T) {
	c, _ := NewLocalCache(10)
	defer c.Close()

	c.Set(ctx, "key", []byte("value"), 50*time.Millisecond)
	time.Sleep(100 * time.Millisecond)

	_, err := c.Get(ctx, "key")
	if err != ErrCacheMiss {
		t.Fatalf("esperaba ErrCacheMiss, got: %v", err)
	}
}

func TestLocalCache_Del(t *testing.T) {
	c, _ := NewLocalCache(10)
	defer c.Close()

	c.Set(ctx, "key", []byte("value"), time.Minute)
	c.Del(ctx, "key")

	_, err := c.Get(ctx, "key")
	if err != ErrCacheMiss {
		t.Fatalf("esperaba ErrCacheMiss después de Del, got: %v", err)
	}
}

func TestLocalCache_LRUEviction(t *testing.T) {
	c, _ := NewLocalCache(3)
	defer c.Close()

	c.Set(ctx, "a", []byte("1"), time.Minute)
	c.Set(ctx, "b", []byte("2"), time.Minute)
	c.Set(ctx, "c", []byte("3"), time.Minute)

	c.Set(ctx, "d", []byte("4"), time.Minute)

	_, err := c.Get(ctx, "a")
	if err != ErrCacheMiss {
		t.Fatal("esperaba que 'a' fuera evictado por LRU")
	}
}

func TestLocalCache_ConcurrentSetGet(t *testing.T) {
	c, _ := NewLocalCache(100)
	defer c.Close()

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(2)

		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			c.Set(ctx, key, []byte("value"), time.Minute)
		}(i)

		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			c.Get(ctx, key)
		}(i)
	}

	wg.Wait()
}

func TestLocalCache_ConcurrentAllOps(t *testing.T) {
	c, _ := NewLocalCache(50)
	defer c.Close()

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(3)

		go func(i int) {
			defer wg.Done()
			c.Set(ctx, fmt.Sprintf("key-%d", i), []byte("v"), time.Minute)
		}(i)

		go func(i int) {
			defer wg.Done()
			c.Get(ctx, fmt.Sprintf("key-%d", i))
		}(i)

		go func(i int) {
			defer wg.Done()
			c.Del(ctx, fmt.Sprintf("key-%d", i))
		}(i)
	}

	wg.Wait()
}

func TestLocalCache_ConcurrentSameKey(t *testing.T) {
	c, _ := NewLocalCache(10)
	defer c.Close()

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(2)

		go func() {
			defer wg.Done()
			c.Set(ctx, "shared", []byte("value"), time.Minute)
		}()

		go func() {
			defer wg.Done()
			c.Get(ctx, "shared")
		}()
	}

	wg.Wait()
}

func TestLocalCache_CleanupLoop(t *testing.T) {
	c, _ := NewLocalCache(100)
	defer c.Close()

	c.ticker.Reset(100 * time.Millisecond)

	for i := range 20 {
		c.Set(ctx, fmt.Sprintf("key-%d", i), []byte("v"), 50*time.Millisecond)
	}

	time.Sleep(300 * time.Millisecond)

	for i := range 20 {
		_, err := c.Get(ctx, fmt.Sprintf("key-%d", i))
		if err != ErrCacheMiss {
			t.Fatalf("key-%d should have be cleanuped", i)
		}
	}
}

func TestLocalCache_StressWithExpiration(t *testing.T) {
	c, _ := NewLocalCache(200)
	defer c.Close()

	var wg sync.WaitGroup
	for i := range 200 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i%20)
			ttl := time.Duration(i%3+1) * 10 * time.Millisecond

			c.Set(ctx, key, []byte("value"), ttl)
			time.Sleep(time.Duration(i%5) * time.Millisecond)
			c.Get(ctx, key)
			c.Del(ctx, key)
		}(i)
	}

	wg.Wait()
}
