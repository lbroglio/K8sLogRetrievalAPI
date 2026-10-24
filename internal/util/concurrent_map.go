package util

import (
	"maps"
	"slices"
	"sync"
)

type ConcurrentMap[K comparable, V any] struct {
	contents map[K]V
	mu       sync.RWMutex
}

// NewConcurrentMap creates a new instance of ConcurrentMap.
func NewConcurrentMap[K comparable, V any]() *ConcurrentMap[K, V] {
	return &ConcurrentMap[K, V]{
		contents: make(map[K]V),
	}
}

// Put adds an item to the ConcurrentMap in a thread-safe manner.
// If the key already exists, it will overwrite the existing value.
func (cm *ConcurrentMap[K, V]) Put(key K, value V) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.contents[key] = value
}

// GetAll returns all values in the ConcurrentMap in a thread-safe manner.
func (cm *ConcurrentMap[K, V]) GetAllValues() []V {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return slices.Collect(maps.Values(cm.contents))
}

// Get returns the value for a given key in the ConcurrentMap in a thread-safe manner.
func (cm *ConcurrentMap[K, V]) Get(key K) V {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.contents[key]
}

// Delete removes an item from the ConcurrentMap in a thread-safe manner.
func (cm *ConcurrentMap[K, V]) Delete(key K) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.contents, key)
}

// ContainsKey checks if a key exists in the ConcurrentMap in a thread-safe manner.
func (cm *ConcurrentMap[K, V]) ContainsKey(key K) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	_, exists := cm.contents[key]
	return exists
}

// PutAll adds all items from another map to the ConcurrentMap in a thread-safe manner.
func (cm *ConcurrentMap[K, V]) PutAll(other map[K]V) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	maps.Copy(cm.contents, other)

}
