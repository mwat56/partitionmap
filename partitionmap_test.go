/*
Copyright © 2024  M.Watermann, 10247 Berlin, Germany

			All rights reserved
		EMail : <support@mwat.de>
*/

package partitionmap

import (
	"reflect"
	"sync"
	"testing"
)

func TestPartitionDel(t *testing.T) {
	// Create a new partition
	partition := &tPartition[int]{
		mtx: &sync.RWMutex{},
		kv:  make(map[string]int),
	}

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
}

func TestPartitionDel_ReturnsNilWhenPartitionIsNil(t *testing.T) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.partition.del(tt.args.aKey)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("del() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPartitionDel_ReturnsSamePartitionWhenKeyDoesNotExist(t *testing.T) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.partition.del(tt.args.aKey)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("del() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPartitionGet_ReturnsEmptyValueAndFalseWhenKeyDoesNotExist(t *testing.T) {
	// Create a new partition
	partition := tPartition[int]{
		mtx: &sync.RWMutex{},
		kv:  make(map[string]int),
	}

	// Test case: key does not exist
	_, ok := partition.get("key1")

	// Verify that the key does not exist
	if ok {
		t.Errorf("Expected key to not exist")
	}
}
