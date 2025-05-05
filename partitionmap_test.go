/*
Copyright © 2024, 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package partitionmap

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

func Test_tPartition_del(t *testing.T) {
	partition := newPartition[int]()

	// Insert a key-value pair
	partition.put("key1", 100)

	// Verify that the key-value pair exists
	if _, ok := partition.kv["key1"]; !ok {
		t.Errorf("Expected key-value pair to exist")
	}

	// Delete the key-value pair
	partition = partition.del("key1")

	// Verify that the key-value pair no longer exists
	if _, ok := partition.kv["key1"]; ok {
		t.Errorf("Expected key-value pair to be deleted")
	}
} // Test_tPartition_del()

func Test_tPartition_del_RemovesKeyValuePair(t *testing.T) {
	partition := newPartition[int]()

	// Add a key-value pair
	partition.put("testKey", 42)

	// Verify key exists
	if val, ok := partition.get("testKey"); !ok || (42 != val) {
		t.Errorf("Expected key 'testKey' to exist with value 42, got %v, exists: %v",
			val, ok)
	}

	// Delete the key
	partition = partition.del("testKey")

	// Verify key no longer exists
	if _, ok := partition.get("testKey"); ok {
		t.Errorf("Expected key 'testKey' to be deleted, but it still exists")
	}
} // Test_tPartition_del_RemovesKeyValuePair()

func Test_tPartition_del_ReturnsNilWhenPartitionIsNil(t *testing.T) {
	type args struct {
		aKey string
	}
	tests := []struct {
		name      string
		partition *tPartition[int]
		args      args
		want      *tPartition[int]
	}{
		{
			name:      "Partition is nil",
			partition: nil,
			args:      args{aKey: "key1"},
			want:      nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.partition.del(tc.args.aKey)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("del() = %v, want %v", got, tc.want)
			}
		})
	}
} // Test_tPartition_del_ReturnsNilWhenPartitionIsNil()

func Test_tPartition_del_ReturnsSamePartitionWhenKeyDoesNotExist(t *testing.T) {
	type args struct {
		aKey string
	}

	tests := []struct {
		name      string
		partition *tPartition[int]
		args      args
		want      *tPartition[int]
	}{
		{
			name:      "Partition is not nil",
			partition: &tPartition[int]{mtx: &sync.RWMutex{}, kv: make(map[string]int)},
			args:      args{aKey: "key1"},
			want:      &tPartition[int]{mtx: &sync.RWMutex{}, kv: make(map[string]int)},
		},
		{
			name:      "Partition is nil",
			partition: nil,
			args:      args{aKey: "key1"},
			want:      nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.partition.del(tc.args.aKey)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("del() = %v, want %v", got, tc.want)
			}
		})
	}
} // Test_tPartition_del_ReturnsSamePartitionWhenKeyDoesNotExist()

func Test_tPartition_get(t *testing.T) {
	partition := newPartition[int]()

	// Add a key-value pair
	partition.put("testKey", 42)

	// Test getting an existing key
	val, ok := partition.get("testKey")
	if !ok || (42 != val) {
		t.Errorf("get() for existing key = %v, %v; want 42, true",
			val, ok)
	}

	// Test getting a non-existent key
	val, ok = partition.get("nonExistentKey")
	if ok || (0 != val) {
		t.Errorf("get() for non-existent key = %v, %v; want 0, false",
			val, ok)
	}

	// // Test with nil partition
	// var nilPartition *tPartition[int]
	// val, ok = nilPartition.get("anyKey")
	// if ok || val != 0 {
	// 	t.Errorf("get() with nil partition = %v, %v; want 0, false", val, ok)
	// }
} // Test_tPartition_get()

func Test_tPartition_get_ReturnsEmptyValueAndFalseWhenKeyDoesNotExist(t *testing.T) {
	partition := newPartition[int]()

	// Test case: key does not exist
	_, ok := partition.get("key1")

	// Verify that the key does not exist
	if ok {
		t.Errorf("Expected key to not exist")
	}
} // Test_tPartition_get_ReturnsEmptyValueAndFalseWhenKeyDoesNotExist()

func Test_tPartition_keys(t *testing.T) {
	partition := newPartition[int]()

	// Test empty partition
	keys := partition.keys()
	if 0 != len(keys) {
		t.Errorf("Expected empty keys slice for empty partition, got %v",
			keys)
	}

	// Add some key-value pairs
	partition.put("key1", 100)
	partition.put("key2", 200)
	partition.put("key3", 300)

	// Get the keys
	keys = partition.keys()

	// Verify the number of keys
	if 3 != len(keys) {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Verify the keys are sorted
	expectedKeys := []string{"key1", "key2", "key3"}
	if !reflect.DeepEqual(keys, expectedKeys) {
		t.Errorf("Expected keys %v, got %v", expectedKeys, keys)
	}
} // Test_tPartition_keys()

func Test_tPartition_put(t *testing.T) {
	partition := newPartition[int]()

	// Test adding a new key-value pair
	partition = partition.put("testKey", 42)

	// Verify the key-value pair was added
	if val, ok := partition.get("testKey"); !ok || (42 != val) {
		t.Errorf("put() failed to add key-value pair, got %v, exists: %v",
			val, ok)
	}

	// Test updating an existing key
	partition = partition.put("testKey", 100)

	// Verify the value was updated
	if val, ok := partition.get("testKey"); !ok || (100 != val) {
		t.Errorf("put() failed to update existing key, got %v, exists: %v",
			val, ok)
	}

	// Test with nil partition
	var nilPartition *tPartition[int]
	result := nilPartition.put("key", 1)
	if nil != result {
		t.Errorf("Expected nil result when calling put() on nil partition")
	}
} // Test_tPartition_put()

func Test_tPartition_String(t *testing.T) {
	partition := newPartition[int]()

	// Test empty partition
	result := partition.String()
	if "" != result {
		t.Errorf("Expected empty string for empty partition, got %q", result)
	}

	// Add some key-value pairs
	partition.put("key1", 100)
	partition.put("key2", 200)
	partition.put("key3", 300)

	// Get the string representation
	result = partition.String()

	// Verify the string contains all keys and values
	for _, key := range []string{"key1", "key2", "key3"} {
		if !strings.Contains(result, key) {
			t.Errorf("Expected string to contain key %q, but it doesn't",
				key)
		}
	}

	for _, val := range []int{100, 200, 300} {
		if !strings.Contains(result, fmt.Sprintf("%d", val)) {
			t.Errorf("Expected string to contain value %d, but it doesn't",
				val)
		}
	}

	// Verify the format matches expected pattern
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if 3 != len(lines) {
		t.Errorf("Expected 3 lines in output, got %d", len(lines))
	}
} // Test_tPartition_String()

func Test_TPartitionMap_partitionIndex(t *testing.T) {
	pm := NewPartitionMap[int]()

	tests := []struct {
		name string
		key  string
	}{
		{"Empty key", ""},
		{"Simple key", "testKey"},
		{"Long key", "this is a very long key with spaces and special chars !@#$%^&*()"},
		{"Duplicate content", "duplicate"},
		{"Same duplicate content", "duplicate"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			idx := pm.partitionIndex(tc.key)

			// Verify index is within valid range
			if idx >= numberOfPartitionsInPartitionMap {
				t.Errorf("partitionIndex(%q) = %d, want < %d",
					tc.key, idx, numberOfPartitionsInPartitionMap)
			}

			// Verify consistency - same key should always produce same index
			idx2 := pm.partitionIndex(tc.key)
			if idx != idx2 {
				t.Errorf("partitionIndex not consistent: %q produced %d and %d",
					tc.key, idx, idx2)
			}
		})
	}

	// Test different keys produce different indices (not guaranteed but likely)
	indices := make(map[uint32]string)
	collisions := 0

	for i := range 100 {
		key := fmt.Sprintf("test-key-%d", i)
		idx := pm.partitionIndex(key)

		if existingKey, exists := indices[idx]; exists {
			collisions++
			t.Logf("Collision detected: %q and %q both map to index %d",
				existingKey, key, idx)
		} else {
			indices[idx] = key
		}
	}

	// Log collision rate - some collisions are expected with CRC32
	t.Logf("Collision rate: %d%% (%d out of 100 keys)", collisions, collisions)
} // Test_TPartitionMap_partitionIndex()

func Test_TPartitionMap_partition(t *testing.T) {
	type tArgs struct {
		aKey    string
		aCreate bool
	}
	tests := []struct {
		name      string
		pm        *TPartitionMap[int]
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
			pm:   NewPartitionMap[int](),
			args: tArgs{
				aKey:    "testKey",
				aCreate: true,
			},
			wantFound: true,
			wantNil:   false,
		},
		{
			name: "Don't create new partition",
			pm:   NewPartitionMap[int](),
			args: tArgs{
				aKey:    "nonExistentKey",
				aCreate: false,
			},
			wantFound: false,
			wantNil:   true,
		},
		{
			name: "Get existing partition",
			pm:   NewPartitionMap[int]().Put("existingKey", 42),
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
				if partition.mtx == nil {
					t.Errorf("partition() returned partition with nil mutex")
				}
				if partition.kv == nil {
					t.Errorf("partition() returned partition with nil map")
				}
			}
		})
	}
} // Test_TPartitionMap_partition()

func Test_TPartitionMap_Delete(t *testing.T) {
	type args struct {
		aKey string
	}
	tests := []struct {
		name string
		pm   *TPartitionMap[int]
		args args
		want bool // whether key should exist after deletion
	}{
		{
			name: "Delete existing key",
			pm: func() *TPartitionMap[int] {
				return NewPartitionMap[int]().Put("testKey", 42)
			}(),
			args: args{aKey: "testKey"},
			want: false,
		},
		{
			name: "Delete non-existent key",
			pm:   NewPartitionMap[int](),
			args: args{aKey: "nonExistentKey"},
			want: false,
		},
		{
			name: "Nil partition map",
			pm:   nil,
			args: args{aKey: "anyKey"},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Store original map for nil check
			originalPM := tc.pm

			// Call Delete method
			result := tc.pm.Delete(tc.args.aKey)

			// Check if result is the same instance as input
			if result != originalPM {
				t.Errorf("Delete() returned different instance, got %v, want %v", result, originalPM)
			}

			// Check if key exists after deletion
			if tc.pm != nil {
				_, exists := tc.pm.Get(tc.args.aKey)
				if exists != tc.want {
					t.Errorf("After Delete(), key existence = %v, want %v", exists, tc.want)
				}
			}
		})
	}
} // Test_TPartitionMap_Delete()

func Test_TPartitionMap_Get(t *testing.T) {
	tests := []struct {
		name      string
		pm        *TPartitionMap[int]
		key       string
		wantValue int
		wantFound bool
	}{
		{
			name:      "Get existing key",
			pm:        NewPartitionMap[int]().Put("testKey", 42),
			key:       "testKey",
			wantValue: 42,
			wantFound: true,
		},
		{
			name:      "Get non-existent key",
			pm:        NewPartitionMap[int](),
			key:       "nonExistentKey",
			wantValue: 0,
			wantFound: false,
		},
		// {
		// 	name:      "Nil partition map",
		// 	pm:        nil,
		// 	args:      tArgs{aKey: "anyKey"},
		// 	wantValue: 0,
		// 	wantFound: false,
		// },
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

func Test_TPartitionMap_Keys(t *testing.T) {
	tests := []struct {
		name string
		pm   *TPartitionMap[int]
		want []string
	}{
		{
			name: "Empty partition map",
			pm:   NewPartitionMap[int](),
			want: []string{},
		},
		{
			name: "Partition map with keys",
			pm: NewPartitionMap[int]().
				Put("key1", 100).
				Put("key2", 200).
				Put("key3", 300),
			want: []string{"key1", "key2", "key3"},
		},
		// {
		// 	name: "Nil partition map",
		// 	pm:   nil,
		// 	want: nil,
		// },
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
		pm   *TPartitionMap[int]
		want int
	}{
		{
			name: "Empty partition map",
			pm:   NewPartitionMap[int](),
			want: 0,
		},
		{
			name: "Partition map with one key",
			pm:   NewPartitionMap[int]().Put("key1", 100),
			want: 1,
		},
		{
			name: "Partition map with multiple keys",
			pm: NewPartitionMap[int]().
				Put("key1", 100).
				Put("key2", 200).
				Put("key3", 300),
			want: 3,
		},
		// {
		// 	name: "Nil partition map",
		// 	pm:   nil,
		// 	want: 0,
		// },
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
		pm        *TPartitionMap[int]
		key       string
		value     int
		wantValue int
		wantFound bool
	}{
		{
			name:      "Add new key-value pair",
			pm:        NewPartitionMap[int](),
			key:       "newKey",
			value:     42,
			wantValue: 42,
			wantFound: true,
		},
		{
			name:      "Update existing key",
			pm:        NewPartitionMap[int]().Put("existingKey", 100),
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

func Test_TPartitionMap_String(t *testing.T) {
	tests := []struct {
		name string
		pm   *TPartitionMap[int]
		want string
	}{
		{
			name: "Empty partition map",
			pm:   NewPartitionMap[int](),
			want: "",
		},
		{
			name: "Partition map with values",
			pm: NewPartitionMap[int]().
				Put("key1", 100).
				Put("key2", 200).
				Put("key3", 300),
			want: "key1: `100`\nkey2: '200'\nkey3: '300'\n", // Assuming this format from tPartition.String()
		},
		// {
		// 	name: "Nil partition map",
		// 	pm:   nil,
		// 	want: "",
		// },
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

/* _EoF_ */
