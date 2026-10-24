package util

import (
	"reflect"
	"slices"
	"sync"
	"testing"
)

func TestNewConcurrentMap_EmptyMap(t *testing.T) {
	concurrentMap := NewConcurrentMap[string, int]()

	if concurrentMap == nil {
		t.Fatal("Expected a concurrent map, got nil")
	}
	if values := concurrentMap.GetAllValues(); len(values) != 0 {
		t.Errorf("Expected an empty concurrent map, got %v", values)
	}
}

func TestConcurrentMap_PutAndGet_Success(t *testing.T) {
	concurrentMap := NewConcurrentMap[string, int]()

	concurrentMap.Put("one", 1)

	if value := concurrentMap.Get("one"); value != 1 {
		t.Errorf("Expected value 1, got %d", value)
	}
	if !concurrentMap.ContainsKey("one") {
		t.Error("Expected concurrent map to contain key one")
	}
}

func TestConcurrentMap_PutExistingKey_OverwritesValue(t *testing.T) {
	concurrentMap := NewConcurrentMap[string, int]()

	concurrentMap.Put("one", 1)
	concurrentMap.Put("one", 2)

	if value := concurrentMap.Get("one"); value != 2 {
		t.Errorf("Expected overwritten value 2, got %d", value)
	}
}

func TestConcurrentMap_GetMissingKey_ReturnsZeroValue(t *testing.T) {
	concurrentMap := NewConcurrentMap[string, int]()

	if value := concurrentMap.Get("missing"); value != 0 {
		t.Errorf("Expected zero value for missing key, got %d", value)
	}
	if concurrentMap.ContainsKey("missing") {
		t.Error("Expected missing key not to be present")
	}
}

func TestConcurrentMap_GetAllValues_ReturnsAllValues(t *testing.T) {
	concurrentMap := NewConcurrentMap[string, int]()
	concurrentMap.Put("one", 1)
	concurrentMap.Put("two", 2)
	concurrentMap.Put("three", 3)

	values := concurrentMap.GetAllValues()
	slices.Sort(values)
	expectedValues := []int{1, 2, 3}

	if !reflect.DeepEqual(values, expectedValues) {
		t.Errorf("Expected values %v, got %v", expectedValues, values)
	}
}

func TestConcurrentMap_Delete_RemovesKey(t *testing.T) {
	concurrentMap := NewConcurrentMap[string, int]()
	concurrentMap.Put("one", 1)

	concurrentMap.Delete("one")

	if concurrentMap.ContainsKey("one") {
		t.Error("Expected deleted key not to be present")
	}
	if value := concurrentMap.Get("one"); value != 0 {
		t.Errorf("Expected zero value after deleting key, got %d", value)
	}
}

func TestConcurrentMap_DeleteMissingKey_Success(t *testing.T) {
	concurrentMap := NewConcurrentMap[string, int]()

	concurrentMap.Delete("missing")

	if values := concurrentMap.GetAllValues(); len(values) != 0 {
		t.Errorf("Expected an empty concurrent map, got %v", values)
	}
}

func TestConcurrentMap_PutAll_AddsAndOverwritesValues(t *testing.T) {
	concurrentMap := NewConcurrentMap[string, int]()
	concurrentMap.Put("one", 1)

	concurrentMap.PutAll(map[string]int{
		"one":   10,
		"two":   2,
		"three": 3,
	})

	values := concurrentMap.GetAllValues()
	slices.Sort(values)
	expectedValues := []int{2, 3, 10}
	if !reflect.DeepEqual(values, expectedValues) {
		t.Errorf("Expected values %v, got %v", expectedValues, values)
	}
	if value := concurrentMap.Get("one"); value != 10 {
		t.Errorf("Expected overwritten value 10, got %d", value)
	}
}

func TestConcurrentMap_GetAllValues_ReturnsSnapshot(t *testing.T) {
	concurrentMap := NewConcurrentMap[string, int]()
	concurrentMap.Put("one", 1)

	values := concurrentMap.GetAllValues()
	values[0] = 2

	if value := concurrentMap.Get("one"); value != 1 {
		t.Errorf("Expected map value to remain 1, got %d", value)
	}
}

func TestConcurrentMap_ConcurrentAccess_Success(t *testing.T) {
	concurrentMap := NewConcurrentMap[int, int]()
	const itemCount = 100

	var waitGroup sync.WaitGroup
	waitGroup.Add(itemCount)
	for i := range itemCount {
		go func() {
			defer waitGroup.Done()
			concurrentMap.Put(i, i)
			if value := concurrentMap.Get(i); value != i {
				t.Errorf("Expected value %d, got %d", i, value)
			}
			if !concurrentMap.ContainsKey(i) {
				t.Errorf("Expected key %d to be present", i)
			}
		}()
	}
	waitGroup.Wait()

	if values := concurrentMap.GetAllValues(); len(values) != itemCount {
		t.Errorf("Expected %d values, got %d", itemCount, len(values))
	}
}
