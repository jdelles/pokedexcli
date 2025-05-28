package pokecache

import (
    "testing"
    "time"
)

func TestAddGet(t *testing.T) {
    cache := NewCache(time.Minute)
    
    cases := []struct {
        key string
        val []byte
    }{
        {
            key: "key1",
            val: []byte("val1"),
        },
        {
            key: "key2",
            val: []byte("val2"),
        },
    }
    
    for _, c := range cases {
        cache.Add(c.key, c.val)
        val, ok := cache.Get(c.key)
        if !ok {
            t.Errorf("expected to find key %s", c.key)
            continue
        }
        if string(val) != string(c.val) {
            t.Errorf("expected value %s, got %s", string(c.val), string(val))
        }
    }
}

func TestReap(t *testing.T) {
    interval := time.Millisecond * 10
    cache := NewCache(interval)
    
    key := "key1"
    val := []byte("val1")
    cache.Add(key, val)
    
    // Verify the key exists
    _, ok := cache.Get(key)
    if !ok {
        t.Errorf("expected to find key")
        return
    }
    
    // Wait for the reaper
    time.Sleep(interval * 2)
    
    // Verify the key was reaped
    _, ok = cache.Get(key)
    if ok {
        t.Errorf("expected key to be reaped")
    }
}

func TestGetMissingKey(t *testing.T) {
    cache := NewCache(time.Minute)
    
    _, ok := cache.Get("missing")
    if ok {
        t.Errorf("expected missing key to return not ok")
    }
}