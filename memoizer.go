package memoizer

import (
	"fmt"
	"strings"
	"sync"

	"golang.org/x/sync/singleflight"
)

// Memoizer is a thread-safe memoizer that caches the results of a function call.
// It is designed to be used in a ephemeral context, such as a http request.
// That's why there's no expiration mechanism on the cache.
type Memoizer struct {
	cache        map[string]memoizedResult
	mutex        sync.RWMutex
	singleflight singleflight.Group
}

// NewMemoizer creates a new memoizer instance.
func NewMemoizer() *Memoizer {
	return &Memoizer{
		cache:        map[string]memoizedResult{},
		singleflight: singleflight.Group{},
	}
}

// Memoize memoizes the result of a function call.
func (m *Memoizer) Memoize(function func() (interface{}, error), keys ...interface{}) (interface{}, error) {
	keySlc := []string{}
	for _, v := range keys {
		keySlc = append(keySlc, fmt.Sprintf("%v", v))
	}
	key := strings.Join(keySlc, "|")

	m.mutex.RLock()
	if res, ok := m.cache[key]; ok {
		m.mutex.RUnlock()
		return res.value, res.err
	}
	m.mutex.RUnlock()

	value, err, _ := m.singleflight.Do(key, func() (interface{}, error) {
		value, err := function()
		m.mutex.Lock()
		m.cache[key] = memoizedResult{value: value, err: err}
		m.mutex.Unlock()
		return value, err
	})
	return value, err
}

type memoizedResult struct {
	value interface{}
	err   error
}
