/*
Copyright © 2024, 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package partitionmap

import (
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_partitionIndex_TypeSpecific(t *testing.T) {
	tests := []struct {
		name     string
		key      any
		wantType string
	}{
		{"Int key", 42, "int"},
		{"Int8 key", int8(-127), "int8"},
		{"Int16 key", int16(-32767), "int16"},
		{"Int32 key", int32(-2147483647), "int32"},
		{"Int64 key", int64(-9223372036854775807), "int64"},
		{"Uint key", uint(42), "uint"},
		{"Uint8 key", uint8(255), "uint8"},
		{"Uint16 key", uint16(65535), "uint16"},
		{"Uint32 key", uint32(4294967295), "uint32"},
		{"Uint64 key", uint64(18446744073709551615), "uint64"},
		{"Uintptr key", uintptr(128), "uintptr"},
		{"Float32 key", float32(3.14), "float32"},
		{"Float64 key", 3.14159, "float64"},
		{"String key", "test key string", "string"},
		{"Custom type key", struct{ name string }{"test struct key"}, "default"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Convert the test key to the generic type
			var idx uint8

			switch tc.wantType {
			case "int":
				idx = partitionIndex(tc.key.(int))
			case "int8":
				idx = partitionIndex(tc.key.(int8))
			case "int16":
				idx = partitionIndex(tc.key.(int16))
			case "int32":
				idx = partitionIndex(tc.key.(int32))
			case "int64":
				idx = partitionIndex(tc.key.(int64))
			case "uint":
				idx = partitionIndex(tc.key.(uint))
			case "uint8":
				idx = partitionIndex(tc.key.(uint8))
			case "uint16":
				idx = partitionIndex(tc.key.(uint16))
			case "uint32":
				idx = partitionIndex(tc.key.(uint32))
			case "uint64":
				idx = partitionIndex(tc.key.(uint64))
			case "uintptr":
				idx = partitionIndex(tc.key.(uintptr))
			case "float32":
				idx = partitionIndex(tc.key.(float32))
			case "float64":
				idx = partitionIndex(tc.key.(float64))
			case "string":
				idx = partitionIndex(tc.key.(string))
			} // switch

			// Verify index is within valid range
			if idx >= numberOfPartitionsInMap {
				t.Errorf("partitionIndex(%v) = %d, want < %d",
					tc.key, idx, numberOfPartitionsInMap)
			}

			// Verify consistency - same key should always produce same index
			var idx2 uint8
			switch tc.wantType {
			case "int":
				idx2 = partitionIndex(tc.key.(int))
			case "int8":
				idx2 = partitionIndex(tc.key.(int8))
			case "int16":
				idx2 = partitionIndex(tc.key.(int16))
			case "int32":
				idx2 = partitionIndex(tc.key.(int32))
			case "int64":
				idx2 = partitionIndex(tc.key.(int64))
			case "uint":
				idx2 = partitionIndex(tc.key.(uint))
			case "uint8":
				idx2 = partitionIndex(tc.key.(uint8))
			case "uint16":
				idx2 = partitionIndex(tc.key.(uint16))
			case "uint32":
				idx2 = partitionIndex(tc.key.(uint32))
			case "uint64":
				idx2 = partitionIndex(tc.key.(uint64))
			case "uintptr":
				idx2 = partitionIndex(tc.key.(uintptr))
			case "float32":
				idx2 = partitionIndex(tc.key.(float32))
			case "float64":
				idx2 = partitionIndex(tc.key.(float64))
			case "string":
				idx2 = partitionIndex(tc.key.(string))
			} // switch

			if idx != idx2 {
				t.Errorf("partitionIndex not consistent: %v produced %d and %d",
					tc.key, idx, idx2)
			}
		})
	}
} // Test_partitionIndex_TypeSpecific()

func Test_TPartitionMap_partition(t *testing.T) {
	type tArgs struct {
		aKey    string
		aCreate bool
	}
	tests := []struct {
		name      string
		pm        *TPartitionMap[string, int]
		args      tArgs
		wantFound bool
		wantNil   bool
	}{
		{
			name: "Nil partition map",
			pm:   nil,
			args: tArgs{
				aKey:    "testKey",
				aCreate: true,
			},
			wantFound: false,
			wantNil:   true,
		},
		{
			name: "Create new partition",
			pm:   New[string, int](),
			args: tArgs{
				aKey:    "testKey",
				aCreate: true,
			},
			wantFound: true,
			wantNil:   false,
		},
		{
			name: "Don't create new partition",
			pm:   New[string, int](),
			args: tArgs{
				aKey:    "nonExistentKey",
				aCreate: false,
			},
			wantFound: false,
			wantNil:   true,
		},
		{
			name: "Get existing partition",
			pm:   New[string, int]().Put("existingKey", 42),
			args: tArgs{
				aKey:    "existingKey",
				aCreate: false,
			},
			wantFound: true,
			wantNil:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			partition, found := tc.pm.partition(tc.args.aKey, tc.args.aCreate)

			if found != tc.wantFound {
				t.Errorf("partition() found = %v, want %v",
					found, tc.wantFound)
			}

			if (nil == partition) != tc.wantNil {
				t.Errorf("partition() nil check failed, got nil: %v, want nil: %v",
					nil == partition, tc.wantNil)
			}

			// Additional check: if we created a partition,
			// verify it's properly initialised.
			if found && !tc.wantNil {
				if nil == partition.kv {
					t.Errorf("partition() returned partition with nil map")
				}
			}
		})
	}
} // Test_TPartitionMap_partition()

func Test_TPartitionMap_Clear(t *testing.T) {
	tests := []struct {
		name string
		pm   *TPartitionMap[string, int]
	}{
		{
			name: "Empty partition map",
			pm:   New[string, int](),
		},
		{
			name: "Partition map with values",
			pm: New[string, int]().
				Put("key1", 100).
				Put("key2", 200).
				Put("key3", 300),
		},
		{
			name: "Nil partition map",
			pm:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Store original map for nil check
			originalPM := tc.pm

			// Call Clear method
			got := tc.pm.Clear()

			// Check if result is the same instance as input
			if got != originalPM {
				t.Errorf("Clear() returned different instance, got %v, want %v",
					got, originalPM)
			}

			// For nil case, verify no changes
			if nil == tc.pm {
				return
			}

			// Verify all key/value pairs were removed
			if 0 != tc.pm.Len() {
				t.Errorf("After Clear(), expected length 0, got %d",
					tc.pm.Len())
			}

			// Verify Keys() returns empty slice
			keys := tc.pm.Keys()
			if 0 != len(keys) {
				t.Errorf("After Clear(), expected empty keys slice, got %v",
					keys)
			}
		})
	}
} // Test_TPartitionMap_Clear()

func Test_TPartitionMap_Delete(t *testing.T) {
	tests := []struct {
		name string
		pm   *TPartitionMap[string, int]
		key  string
	}{
		{
			name: "Delete existing key",
			pm: func() *TPartitionMap[string, int] {
				return New[string, int]().Put("testKey", 42)
			}(),
			key: "testKey",
		},
		{
			name: "Delete non-existent key",
			pm:   New[string, int](),
			key:  "nonExistentKey",
		},
		{
			name: "Nil partition map",
			pm:   nil,
			key:  "anyKey",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Store original map for nil check
			originalPM := tc.pm

			// Call Delete method
			got := tc.pm.Delete(tc.key)

			// Check if result is the same instance as input
			if got != originalPM {
				t.Errorf("Delete() returned different instance, got %v, want %v",
					got, originalPM)
			}

			// Check if key exists after deletion
			if tc.pm != nil {
				if _, exists := tc.pm.Get(tc.key); exists {
					t.Errorf("After Delete(), key existence = %q, want %v",
						tc.key, false)
				}
			}
		})
	}
} // Test_TPartitionMap_Delete()

func Test_TPartitionMap_ForEach(t *testing.T) {
	tests := []struct {
		name string
		pm   *TPartitionMap[string, int]
	}{
		{
			name: "Empty partition map",
			pm:   New[string, int](),
		},
		{
			name: "Partition map with values",
			pm: New[string, int]().
				Put("key1", 100).
				Put("key2", 200).
				Put("key3", 300),
		},
		{
			name: "Nil partition map",
			pm:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a map to store visited key-value pairs
			visited := make(map[string]int)

			// Call ForEach to collect all key-value pairs
			got := tc.pm.ForEach(func(key string, value int) {
				visited[key] = value
			})

			// Verify the result is the same instance as input
			if got != tc.pm {
				t.Errorf("ForEach() returned different instance, got %v, want %v",
					got, tc.pm)
			}

			// For nil case, verify no function calls occurred
			if tc.pm == nil {
				if 0 < len(visited) {
					t.Errorf("ForEach() should not call function for nil partition map")
				}
				return
			}

			// For non-nil case, verify all key-value pairs were visited
			keys := tc.pm.Keys()
			if len(visited) != len(keys) {
				t.Errorf("Expected %d key-value pairs to be visited, got %d",
					len(keys), len(visited))
			}

			// Verify each key-value pair was visited correctly
			for _, key := range keys {
				expectedValue, _ := tc.pm.Get(key)
				if value, ok := visited[key]; !ok || value != expectedValue {
					t.Errorf("Expected key %q with value %d, got %v, exists: %v",
						key, expectedValue, value, ok)
				}
			}
		})
	}
} // Test_TPartitionMap_ForEach()

func Test_TPartitionMap_Get(t *testing.T) {
	tests := []struct {
		name      string
		pm        *TPartitionMap[string, int]
		key       string
		wantValue int
		wantFound bool
	}{
		{
			name:      "Get existing key",
			pm:        New[string, int]().Put("testKey", 42),
			key:       "testKey",
			wantValue: 42,
			wantFound: true,
		},
		{
			name:      "Get non-existent key",
			pm:        New[string, int](),
			key:       "nonExistentKey",
			wantValue: 0,
			wantFound: false,
		},
		{
			name:      "Nil partition map",
			pm:        nil,
			key:       "anyKey",
			wantValue: 0,
			wantFound: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotValue, gotFound := tc.pm.Get(tc.key)
			if gotValue != tc.wantValue {
				t.Errorf("Get() value = %v, want %v",
					gotValue, tc.wantValue)
			}
			if gotFound != tc.wantFound {
				t.Errorf("Get() found = %v, want %v",
					gotFound, tc.wantFound)
			}
		})
	}
} // Test_TPartitionMap_Get()

func Test_TPartitionMap_GetOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		pm       *TPartitionMap[string, int]
		key      string
		defValue int
		want     int
	}{
		{
			name:     "Get existing key",
			pm:       New[string, int]().Put("testKey", 42),
			key:      "testKey",
			defValue: -1,
			want:     42,
		},
		{
			name:     "Get non-existent key",
			pm:       New[string, int](),
			key:      "nonExistentKey",
			defValue: 99,
			want:     99,
		},
		{
			name:     "Nil partition map",
			pm:       nil,
			key:      "anyKey",
			defValue: 123,
			want:     123,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.pm.GetOrDefault(tc.key, tc.defValue)
			if got != tc.want {
				t.Errorf("GetOrDefault() = '%v', want '%v'",
					got, tc.want)
			}
		})
	}
} // Test_TPartitionMap_GetOrDefault()

func Test_TPartitionMap_Keys(t *testing.T) {
	tests := []struct {
		name string
		pm   *TPartitionMap[string, int]
		want []string
	}{
		{
			name: "Empty partition map",
			pm:   New[string, int](),
			want: []string{},
		},
		{
			name: "Partition map with keys",
			pm: New[string, int]().
				Put("key1", 100).
				Put("key2", 200).
				Put("key3", 300),
			want: []string{"key1", "key2", "key3"},
		},
		{
			name: "Nil partition map",
			pm:   nil,
			want: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.pm.Keys()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Keys() = %v, want %v",
					got, tc.want)
			}
		})
	}
} // Test_TPartitionMap_Keys()

func Test_TPartitionMap_Len(t *testing.T) {
	tests := []struct {
		name string
		pm   *TPartitionMap[string, int]
		want int
	}{
		{
			name: "Empty partition map",
			pm:   New[string, int](),
			want: 0,
		},
		{
			name: "Partition map with one key",
			pm:   New[string, int]().Put("key1", 100),
			want: 1,
		},
		{
			name: "Partition map with multiple keys",
			pm: New[string, int]().
				Put("key1", 100).
				Put("key2", 200).
				Put("key3", 300),
			want: 3,
		},
		{
			name: "Nil partition map",
			pm:   nil,
			want: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.pm.Len()
			if got != tc.want {
				t.Errorf("Len() = %v, want %v",
					got, tc.want)
			}
		})
	}
} // Test_TPartitionMap_Len()

func Test_TPartitionMap_Put(t *testing.T) {
	tests := []struct {
		name      string
		pm        *TPartitionMap[string, int]
		key       string
		value     int
		wantValue int
		wantFound bool
	}{
		{
			name:      "Add new key-value pair",
			pm:        New[string, int](),
			key:       "newKey",
			value:     42,
			wantValue: 42,
			wantFound: true,
		},
		{
			name:      "Update existing key",
			pm:        New[string, int]().Put("existingKey", 100),
			key:       "existingKey",
			value:     200,
			wantValue: 200,
			wantFound: true,
		},
		{
			name:      "Nil partition map",
			pm:        nil,
			key:       "anyKey",
			value:     42,
			wantValue: 0,
			wantFound: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Store original map for nil check
			originalPM := tc.pm

			// Call Put method
			result := tc.pm.Put(tc.key, tc.value)

			// Check if result is the same instance as input
			if result != originalPM {
				t.Errorf("Put() returned different instance, got %v, want %v",
					result, originalPM)
			}

			// Check if value was stored correctly (except for nil case)
			if nil != tc.pm {
				gotValue, gotFound := tc.pm.Get(tc.key)
				if gotValue != tc.wantValue {
					t.Errorf("After Put(), value = %v, want %v",
						gotValue, tc.wantValue)
				}
				if gotFound != tc.wantFound {
					t.Errorf("After Put(), found = %v, want %v",
						gotFound, tc.wantFound)
				}
			}
		})
	}
} // Test_TPartitionMap_Put()

func Test_TPartitionMap_StressTest_Int64Keys(t *testing.T) {
	// This test is designed to:
	//
	// 1. Run multiple goroutines concurrently to test thread safety with int64 keys.
	// 2. Exercise all `TPartitionMap` methods with int64 keys and string values.
	// 3. Perform a mix of operations to simulate real-world usage patterns.
	// 4. Verify the map remains functional after stress testing.
	// 5. Report performance metrics (operations per second).
	//
	// The test can be skipped in short mode with `go test -short` for
	// quicker test runs.
	if testing.Short() {
		t.Skip("Skipping int64 keys stress test in short mode")
	}

	const (
		numOperations = 1 << 16
		numGoroutines = 1 << 8
	)

	pm := New[int64, string]()
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Create a channel to coordinate operations
	ops := make(chan int, numOperations)
	for i := range numOperations {
		ops <- i
	}
	close(ops)

	// Create a function to run operations concurrently
	runOperations := func(id int) {
		defer wg.Done()
		for op := range ops {
			// Use int64 keys
			key := int64(id*numOperations + op)
			value := fmt.Sprintf("value-%d-%d", id, op)

			// Mix of operations to stress test all capabilities
			switch op % 8 {
			case 0: // Put
				pm.Put(key, value)
			case 1: // Get
				pm.Get(key)
			case 2: // Delete
				pm.Delete(key)
			case 3: // ForEach (limited scope)
				pm.ForEach(func(k int64, v string) {
					// Just access the values
					_ = fmt.Sprintf("%d:%s", k, v)
				})
			case 4: // Keys
				keys := pm.Keys()
				if 100 < len(keys) {
					// Limit memory usage in test
					keys = keys[:100]
					_ = keys
				}
			case 5: // Len
				_ = pm.Len()
			case 6: // String
				s := pm.String()
				if 0 < len(s) {
					// Limit memory usage in test
					s = ""
					_ = s
				}
			case 7: // Values
				values := pm.Values()
				if 100 < len(values) {
					// Limit memory usage in test
					values = values[:100]
					_ = values
				}
			} // switch op % 8

			// Occasionally perform more expensive operations
			if 0 == op%1024 {
				// Get length
				_ = pm.Len()
			}

			// Very occasionally clear the map
			if (0 == op%512) && (1 == id) {
				pm.Clear()
			}
		}
	}

	// Start goroutines
	start := time.Now()
	for i := range numGoroutines {
		go runOperations(i)
	}

	// Wait for all operations to complete
	wg.Wait()
	duration := time.Since(start)

	// Report results
	t.Logf("Int64 keys stress test completed: %d operations across %d goroutines in %v",
		numOperations, numGoroutines, duration)
	t.Logf("Operations per second: %.2f", float64(numOperations)/duration.Seconds())
	t.Logf("Final map size: %d entries", pm.Len())

	// Verify the map is still functional after stress
	testKey := int64(9223372036854775807) // max int64
	testValue := "max-int64-test-value"
	pm.Put(testKey, testValue)
	val, exists := pm.Get(testKey)
	if !exists || (val != testValue) {
		t.Errorf("Map integrity check failed after int64 keys stress test")
	}

	metrics := pm.PartitionStats()
	t.Logf("StressTest_Int64Keys(): partitions = %d, keys = %d, average = %d, metrics.PartKeys = %v",
		metrics.Parts, metrics.Keys, metrics.Avg, metrics.PartKeys)
} // Test_TPartitionMap_StressTest_Int64Keys()

func Test_TPartitionMap_StressTest_StringKeys(t *testing.T) {
	// This test is designed to:
	//
	// 1. Run multiple goroutines concurrently to test thread safety.
	// 2. Exercise all `TPartitionMap` methods: `Put()`, `Get()`, `Delete()`,
	// `ForEach()`, `Keys()`, `Len()`, `String()`, `Clear()`.
	// 3. Perform a mix of operations to simulate real-world usage patterns.
	// 4. Verify the map remains functional after stress testing.
	// 5. Report performance metrics (operations per second).
	//
	// The test can be skipped in short mode with `go test -short` for
	// quicker test runs.
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	const (
		numOperations = 1 << 16
		numGoroutines = 1 << 8
	)

	pm := New[string, string]()
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Create a channel to coordinate operations
	ops := make(chan int, numOperations)
	for i := range numOperations {
		ops <- i
	}
	close(ops)

	// Create a function to run operations concurrently
	runOperations := func(id int) {
		defer wg.Done()
		for op := range ops {
			key := fmt.Sprintf("key-%d-%d", id, op)
			value := fmt.Sprintf("value-%d-%d", id, op)

			// Mix of operations to stress test all capabilities
			switch op % 8 {
			case 0: // Put
				pm.Put(key, value)
			case 1: // Get
				pm.Get(key)
			case 2: // Delete
				pm.Delete(key)
			case 3: // ForEach (limited scope)
				pm.ForEach(func(k string, v string) {
					// Just access the values
					_ = k + v
				})
			case 4: // Keys
				keys := pm.Keys()
				if 100 < len(keys) {
					// Limit memory usage in test
					keys = keys[:100]
					_ = keys
				}
			case 5: // Len
				_ = pm.Len()
			case 6: // String
				s := pm.String()
				if 0 < len(s) {
					// Limit memory usage in test
					s = ""
					_ = s
				}
			case 7: // Values
				values := pm.Values()
				if 100 < len(values) {
					// Limit memory usage in test
					values = values[:100]
					_ = values
				}
			} // switch op % 8

			// Occasionally perform more expensive operations
			if 0 == op%1024 {
				// Get length
				_ = pm.Len()
			}

			// Very occasionally clear the map
			if (0 == op%512) && (1 == id) {
				pm.Clear()
			}
		}
	}

	// Start goroutines
	start := time.Now()
	for i := range numGoroutines {
		go runOperations(i)
	}

	// Wait for all operations to complete
	wg.Wait()
	duration := time.Since(start)

	// Report results
	t.Logf("Stress test completed: %d operations across %d goroutines in %v",
		numOperations, numGoroutines, duration)
	t.Logf("Operations per second: %.2f", float64(numOperations)/duration.Seconds())
	t.Logf("Final map size: %d entries", pm.Len())

	// Verify the map is still functional after stress
	testKey := "final-test-key"
	testValue := "final-test-value"
	pm.Put(testKey, testValue)
	val, exists := pm.Get(testKey)
	if !exists || (val != testValue) {
		t.Errorf("Map integrity check failed after stress test")
	}

	metrics := pm.PartitionStats()
	t.Logf("StressTest_StringKeys(): partitions = %d, keys = %d, average = %d, metrics.PartKeys = %v",
		metrics.Parts, metrics.Keys, metrics.Avg, metrics.PartKeys)
} // Test_TPartitionMap_StressTest_StringKeys()

func Test_TPartitionMap_String(t *testing.T) {
	tests := []struct {
		name string
		pm   *TPartitionMap[string, int]
		want string
	}{
		{
			name: "Empty partition map",
			pm:   New[string, int](),
			want: "",
		},
		{
			name: "Partition map with values",
			pm: New[string, int]().
				Put("key1", 100).
				Put("key2", 200).
				Put("key3", 300),
			want: "key1: `100`\nkey2: '200'\nkey3: '300'\n", // Assuming this format from tPartition.String()
		},
		{
			name: "Nil partition map",
			pm:   nil,
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.pm.String()
			// Since we can't guarantee the order of keys in the output,
			// we'll check that all expected key-value pairs are present
			if (nil == tc.pm) || (0 == tc.pm.Len()) {
				if got != tc.want {
					t.Errorf("String() = %q, want %q",
						got, tc.want)
				}
			} else {
				for _, key := range []string{"key1", "key2", "key3"} {
					if !strings.Contains(got, key) {
						t.Errorf("String() = %q, should contain key %q",
							got, key)
					}
				}
				for _, val := range []int{100, 200, 300} {
					valStr := fmt.Sprintf("%d", val)
					if !strings.Contains(got, valStr) {
						t.Errorf("String() = %q, should contain value %q",
							got, valStr)
					}
				}
			}
		})
	}
} // Test_TPartitionMap_String()

func Test_TPartitionMap_Values(t *testing.T) {
	tests := []struct {
		name      string
		pm        *TPartitionMap[string, int]
		wantCount int
		wantVals  []int
	}{
		{
			name:      "Empty partition map",
			pm:        New[string, int](),
			wantCount: 0,
			wantVals:  []int{},
		},
		{
			name: "Partition map with values",
			pm: New[string, int]().
				Put("key1", 100).
				Put("key2", 200).
				Put("key3", 300),
			wantCount: 3,
			wantVals:  []int{100, 200, 300},
		},
		{
			name:      "Nil partition map",
			pm:        nil,
			wantCount: 0,
			wantVals:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.pm.Values()

			// Check length
			if len(got) != tc.wantCount {
				t.Errorf("Values() returned %d values, want %d",
					len(got), tc.wantCount)
			}

			// For nil case, verify nil return
			if nil == tc.pm {
				if nil != got {
					t.Errorf("Values() returned non-nil for nil map, got %v",
						got)
				}
				return
			}

			// For non-nil case with values, verify all expected values are present
			// (order may vary based on key sorting)
			if 0 < tc.wantCount {
				// Sort both slices to ensure consistent comparison
				sortedGot := make([]int, len(got))
				copy(sortedGot, got)
				sort.Ints(sortedGot)

				sortedWant := make([]int, len(tc.wantVals))
				copy(sortedWant, tc.wantVals)
				sort.Ints(sortedWant)
				if !slices.Equal(sortedGot, sortedWant) {
					// if !reflect.DeepEqual(sortedGot, sortedWant) {
					t.Errorf("Values() = %v, want %v",
						sortedGot, sortedWant)
				}
			}
		})
	}
} // Test_TPartitionMap_Values()

func Test_TPartitionMap_PartitionStats(t *testing.T) {
	tests := []struct {
		name             string
		pm               *TPartitionMap[string, int]
		wantUsedCount    int
		wantNonEmptyKeys int
	}{
		{
			name:             "Empty partition map",
			pm:               New[string, int](),
			wantUsedCount:    0,
			wantNonEmptyKeys: 0,
		},
		{
			name:             "Partition map with one key",
			pm:               New[string, int]().Put("key1", 100),
			wantUsedCount:    1,
			wantNonEmptyKeys: 1,
		},
		{
			name: "Partition map with multiple keys",
			pm: New[string, int]().
				Put("key1", 100).
				Put("key2", 200).
				Put("key3", 300),
			wantUsedCount:    0, // This will vary based on hash distribution
			wantNonEmptyKeys: 3,
		},
		{
			name:             "Nil partition map",
			pm:               nil,
			wantUsedCount:    0,
			wantNonEmptyKeys: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			metrics := tc.pm.PartitionStats()

			// For nil case, verify nil return
			if nil == tc.pm {
				if nil != metrics {
					t.Errorf("PartitionStats() returned non-nil metrics for nil partition map")
				}
				return
			}

			// For the "multiple keys" test, we can't predict exactly how
			// many partitions will be used due to the hash function, so
			// we skip the exact count check
			if ("Partition map with multiple keys" != tc.name) &&
				(metrics.Parts != tc.wantUsedCount) {
				t.Errorf("PartitionStats() metrics.Parts = %v, want %v",
					metrics.Parts, tc.wantUsedCount)
			}

			// Count total keys across all partitions
			totalKeys := 0
			for _, count := range metrics.PartKeys {
				totalKeys += count
			}

			if totalKeys != tc.wantNonEmptyKeys {
				t.Errorf("PartitionStats() total keys = %v, want %v",
					totalKeys, tc.wantNonEmptyKeys)
			}

			// Verify that the number of used partitions matches the map length
			if len(metrics.PartKeys) != metrics.Parts {
				t.Errorf("PartitionStats() map length = %v, metrics.Parts = %v",
					len(metrics.PartKeys), metrics.Parts)
			}

			// Verify that the reported partition counts match actual counts
			if tc.pm != nil {
				for idx, count := range metrics.PartKeys {
					if idx < 0 || idx >= numberOfPartitionsInMap {
						t.Errorf("PartitionStats() invalid partition index: %d", idx)
					}

					// Verify the partition exists and has the reported number of keys
					partition := tc.pm.tPartitionList[idx]
					if partition == nil {
						t.Errorf("PartitionStats() reported non-nil partition at index %d, but it's nil", idx)
					} else if partition.len() != count {
						t.Errorf("PartitionStats() reported %d keys for partition %d, but it has %d",
							count, idx, partition.len())
					}
				}
			}
		})
	}
} // Test_TPartitionMap_PartitionStats()

/* _EoF_ */
